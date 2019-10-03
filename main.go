package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/common/responses"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
	"github.com/grupokindynos/plutus/controllers"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/joho/godotenv/autoload"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
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
		walletsCtrl := controllers.WalletController{}
		api.GET("/status/:coin", func(context *gin.Context) { VerifyRequest(context, walletsCtrl.GetNodeStatus) })
		api.GET("/info/:coin", func(context *gin.Context) { VerifyRequest(context, walletsCtrl.GetInfo) })
		api.GET("/balance/:coin", func(context *gin.Context) { VerifyRequest(context, walletsCtrl.GetWalletInfo) })
		api.GET("/tx/:coin/:txid", func(context *gin.Context) { VerifyRequest(context, walletsCtrl.GetTx) })
		api.GET("/address/:coin", func(context *gin.Context) { VerifyRequest(context, walletsCtrl.GetAddress) })
		api.POST("/validate/address", func(context *gin.Context) { VerifyRequest(context, walletsCtrl.ValidateAddress) })
		api.POST("/send/address", func(context *gin.Context) { VerifyRequest(context, walletsCtrl.SendToAddress) })
		api.POST("/send/cold", func(context *gin.Context) { VerifyRequest(context, walletsCtrl.SendToColdStorage) })
		api.POST("/send/exchange", func(context *gin.Context) { VerifyRequest(context, walletsCtrl.SendToExchange) })
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
	header, body, err := mrt.CreateMRTToken("plutus", os.Getenv("MASTER_PASSWORD"), response, os.Getenv("PLUTUS_PRIVATE_KEY"))
	if err != nil {
		responses.GlobalResponseError(nil, err, c)
		return
	}
	responses.GlobalResponseMRT(header, body, c)
	return
}
