package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/grupokindynos/adrestia-go/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/plutus"
	"github.com/grupokindynos/common/responses"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
	"github.com/grupokindynos/plutus/controllers"
	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

type CurrentTime struct {
	Hour   int
	Day    int
	Minute int
	Second int
}

var currTime CurrentTime

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	currTime = CurrentTime{
		Hour:   time.Now().Hour(),
		Day:    time.Now().Day(),
		Minute: time.Now().Minute(),
		Second: time.Now().Second(),
	}
	App := GetApp()
	_ = App.Run(":" + port)

}

func GetApp() *gin.Engine {
	App := gin.Default()
	App.Use(cors.Default())
	ApplyRoutes(App)
	return App
}

func ApplyRoutes(r *gin.Engine) {
	authUser := os.Getenv("PLUTUS_AUTH_USERNAME")
	authPassword := os.Getenv("PLUTUS_AUTH_PASSWORD")
	api := r.Group("/", gin.BasicAuth(gin.Accounts{
		authUser: authPassword,
	}))
	{
		ctrl := controllers.NewPlutusController()
		go timer(ctrl)
		api.GET("/balance/:coin", func(context *gin.Context) { VerifyRequest(context, ctrl.GetBalance) })
		api.GET("/address/:coin", func(context *gin.Context) { VerifyRequest(context, ctrl.GetAddress) })
		api.POST("/validate/addr", func(context *gin.Context) { VerifyRequest(context, ctrl.ValidateAddress) })
		api.POST("/validate/tx", func(context *gin.Context) { VerifyRequest(context, ctrl.ValidateRawTx) })
		api.POST("/send/address", func(context *gin.Context) { VerifyRequest(context, ctrl.SendToAddress) })
	}
	apiV2 := r.Group("/v2/", gin.BasicAuth(gin.Accounts{
		authUser: authPassword,
	}))
	{
		ctrlV2 := controllers.NewPlutusControllerV2()
		apiV2.GET("/balance/:coin", func(context *gin.Context) { VerifyRequestV2(context, ctrlV2.GetBalanceV2) })
		apiV2.GET("/address/:coin", func(context *gin.Context) { VerifyRequestV2(context, ctrlV2.GetAddressV2) })
		apiV2.POST("/validate/addr", func(context *gin.Context) { VerifyRequestV2(context, ctrlV2.ValidateAddressV2) })
		apiV2.POST("/validate/tx", func(context *gin.Context) { VerifyRequestV2(context, ctrlV2.ValidateRawTxV2) })
		apiV2.POST("/send/address", func(context *gin.Context) { VerifyRequestV2(context, ctrlV2.SendToAddressV2) })
	}
	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Not Found")
	})
}

func VerifyRequest(c *gin.Context, method func(params controllers.Params) (interface{}, error)) {
	payload, err := mvt.VerifyRequest(c)
	if err != nil {
		responses.GlobalResponseNoAuth(c)
		return
	}
	params := controllers.Params{
		Coin: c.Param("coin"),
		Txid: c.Param("txid"),
		Body: payload,
	}
	response, err := method(params)
	if err != nil {
		responses.GlobalResponseError(nil, err, c)
		return
	}
	header, body, err := mrt.CreateMRTToken("plutus", os.Getenv("MASTER_PASSWORD"), response, os.Getenv("PLUTUS_PRIVATE_KEY"))
	if err != nil {
		responses.GlobalResponseError(nil, err, c)
		return
	}
	responses.GlobalResponseMRT(header, body, c)
	return
}

func VerifyRequestV2(c *gin.Context, method func(params controllers.ParamsV2) (interface{}, error)) {
	payload, err := mvt.VerifyRequest(c)
	if err != nil {
		responses.GlobalResponseNoAuth(c)
		return
	}
	variables := c.Request.URL.Query()
	params := controllers.ParamsV2{
		Coin:    c.Param("coin"),
		Txid:    c.Param("txid"),
		Body:    payload,
		Service: variables.Get("source"),
	}
	response, err := method(params)
	if err != nil {
		responses.GlobalResponseError(nil, err, c)
		return
	}
	header, body, err := mrt.CreateMRTToken("plutus", os.Getenv("MASTER_PASSWORD"), response, os.Getenv("PLUTUS_PRIVATE_KEY"))
	if err != nil {
		responses.GlobalResponseError(nil, err, c)
		return
	}
	responses.GlobalResponseMRT(header, body, c)
	return
}

func timer(ctrl *controllers.Controller) {
	for {
		time.Sleep(1 * time.Second)
		currTime = CurrentTime{
			Hour:   time.Now().Hour(),
			Day:    time.Now().Day(),
			Minute: time.Now().Minute(),
			Second: time.Now().Second(),
		}
		if currTime.Second == 0 {
			var wg sync.WaitGroup
			wg.Add(1)
			runCrons(&wg, ctrl)
			wg.Wait()
		}
	}
}

func runCrons(mainWg *sync.WaitGroup, ctrl *controllers.Controller) {
	defer func() {
		mainWg.Done()
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go runCronMinutes(180, runSend, &wg, ctrl) // 180 minutes
	wg.Wait()
}

func runCronMinutes(schedule int, function func(ctrl *controllers.Controller), wg *sync.WaitGroup, ctrl *controllers.Controller) {
	go func() {
		defer func() {
			wg.Done()
		}()
		remainder := currTime.Minute % schedule
		if remainder == 0 {
			function(ctrl)
		}
		return
	}()
}

var floatType = reflect.TypeOf(float64(0))

func runSend(ctrl *controllers.Controller) {
	fmt.Println("Running send script")
	for _, coin := range coinfactory.Coins {
		if coin.Info.StableCoin || coin.Info.Tag == "DASH" {
			continue
		}
		params := controllers.Params{
			Coin: coin.Info.Tag,
		}
		balance, err := ctrl.GetBalance(params)
		if err != nil {
			continue
		}
		plutusBalance, ok := balance.(plutus.Balance)
		if !ok {
			continue
		}
		send := checkSend(plutusBalance.Confirmed, coin.Info.Tag)
		if send {
			address, err := getDepositAddress(coin.Info.Tag)
			if address == "" || err != nil {
				continue
			}
			sendInfo := plutus.SendAddressBodyReq{
				Amount:  plutusBalance.Confirmed,
				Address: address,
				Coin:    coin.Info.Tag,
			}
			rawData, err := json.Marshal(sendInfo)
			if err != nil {
				continue
			}
			newParams := controllers.Params{
				Coin: coin.Info.Tag,
				Body: rawData,
			}
			txId, err := ctrl.SendToAddress(newParams)
			log.Println("sent ", sendInfo.Amount, " ", sendInfo.Coin, " to ", sendInfo.Address)
			fmt.Println(txId, err)
		}
	}
}

func checkSend(balance float64, coin string) bool {
	if coin != "BTC" && coin != "LTC" && balance > 1 {
		return true
	}
	if coin == "BTC" && balance > 0.001 {
		return true
	}
	if coin == "LTC" && balance > 0.1 {
		return true
	}
	return false
}

func getDepositAddress(coin string) (string, error) {
	addr, err := getAddress(coin, "http://adrestia.polispay.com")
	return addr.ExchangeAddress.Address, err
}

func getAddress(coin string, adrestiaUrl string) (address models.AddressResponse, err error) {
	url := adrestiaUrl + "/address/" + coin
	req, err := mvt.CreateMVTToken("GET", url, "plutus", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("PLUTUS_PRIVATE_KEY"))
	if err != nil {
		return
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		err = errors.New("no header signature")
		return
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("ADRESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return
	}
	err = json.Unmarshal(payload, &address)
	if err != nil {
		return
	}
	return
}
