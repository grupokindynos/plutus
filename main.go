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
		port = "8081"
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
		api.GET("/balance/:coin", func(context *gin.Context) { VerifyRequest(context, ctrl.GetBalance) })
		api.GET("/address/:coin", func(context *gin.Context) { VerifyRequest(context, ctrl.GetAddress) })
		api.POST("/validate/addr", func(context *gin.Context) { VerifyRequest(context, ctrl.ValidateAddress) })
		api.POST("/validate/tx", func(context *gin.Context) { VerifyRequest(context, ctrl.ValidateRawTx) })
		api.POST("/send/address", func(context *gin.Context) { VerifyRequest(context, ctrl.SendToAddress) })
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
