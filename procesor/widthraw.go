package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
	"github.com/grupokindynos/common/plutus"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"time"

	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/plutus/controllers"
	_ "github.com/joho/godotenv/autoload"
)

type AdrestiaRequests struct {
	AdrestiaUrl string
}

var adrestia = AdrestiaRequests{
	AdrestiaUrl: "http://adrestia.polispay.com/",
}

var floatType = reflect.TypeOf(float64(0))

func main() {
	ctrl := controllers.NewPlutusController()
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
				Amount: plutusBalance.Confirmed,
				Address: address,
				Coin: coin.Info.Tag,
			}
			rawData, err := json.Marshal(sendInfo)
			if err != nil {
				continue
			}
			newParams := controllers.Params{
				Coin: coin.Info.Tag,
				Body: rawData,
			}
			txid, err := ctrl.SendToAddress(newParams)
			fmt.Println(txid, err)
		}
	}
}

func checkSend(balance float64, coin string) bool {
	if coin != "BTC" && balance > 1 {
		return true
	}
	if coin == "BTC" && balance > 0.001 {
		return true
	}
	return false
}

func getDepositAddress(coin string) (string, error) {
	addr, err := adrestia.GetAddress(coin)
	return addr.ExchangeAddress.Address, err
}

func (a *AdrestiaRequests) GetAddress(coin string) (address models.AddressResponse, err error) {
	url := a.AdrestiaUrl + "address/" + coin
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
