package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/plutus/controllers/wallets"
	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/joho/godotenv"
	"net/http"
	"os"
)

func init() {
	_ = godotenv.Load()
}

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
		walletsCtrl := wallets.WalletController{}
		api.GET(":coin/status", walletsCtrl.GetNodeStatus)
		api.GET(":coin/info", walletsCtrl.GetInfo)
		api.GET(":coin/balance", walletsCtrl.GetWalletInfo)
		api.GET(":coin/address", walletsCtrl.GetAddress)
		api.GET(":coin/send/address/:address", walletsCtrl.SendToAddress)
		api.GET(":coin/send/cold", walletsCtrl.SendToColdStorage)
		api.GET(":coin/send/exchange", walletsCtrl.SendToExchange)

	}
	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Not Found")
	})
}
