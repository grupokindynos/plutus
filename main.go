package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/plutus/controllers"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/joho/godotenv/autoload"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
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
	authUser := os.Getenv("AUTH_USERNAME")
	authPassword := os.Getenv("AUTH_PASSWORD")
	api := r.Group("/", gin.BasicAuth(gin.Accounts{
		authUser: authPassword,
	}))
	{
		walletsCtrl := controllers.WalletController{}
		api.GET("/status/:coin", walletsCtrl.GetNodeStatus)
		api.GET("/info/:coin", walletsCtrl.GetInfo)
		api.GET("/balance/:coin", walletsCtrl.GetWalletInfo)
		api.GET("/tx/:coin/:txid", walletsCtrl.GetTx)
		api.GET("/address/:coin", walletsCtrl.GetAddress)
		api.GET("/send/address/:coin/:address", walletsCtrl.SendToAddress)
		api.GET("/send/cold/:coin", walletsCtrl.SendToColdStorage)
		api.GET("/send/exchange/:coin", walletsCtrl.SendToExchange)
		api.GET("/validate/address/:coin/:address", walletsCtrl.ValidateAddress)

	}
	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Not Found")
	})
}
