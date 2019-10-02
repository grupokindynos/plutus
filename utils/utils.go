package utils

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/common/jwt"
	"github.com/grupokindynos/common/tokens/mvt"
	"github.com/grupokindynos/plutus/config"
	"os"
)

func VerifyRequest(c *gin.Context) (payload []byte, err error) {
	reqBody, _ := c.GetRawData()
	headerSignature := c.GetHeader("service")
	if headerSignature == "" {
		return nil, config.ErrorNoHeaderSignature
	}
	decodedHeader, err := jwt.DecodeJWSNoVerify(headerSignature)
	if err != nil {
		return nil, config.ErrorSignatureParse
	}
	var serviceStr string
	err = json.Unmarshal(decodedHeader, &serviceStr)
	if err != nil {
		return nil, config.ErrorUnmarshal
	}
	// Check which service the request is announcing
	var pubKey string
	switch serviceStr {
	case "ladon":
		pubKey = os.Getenv("LADON_PUBLIC_KEY")
	case "tyche":
		pubKey = os.Getenv("TYCHE_PUBLIC_KEY")
	case "adrestia":
		pubKey = os.Getenv("ADRESTIA_PUBLIC_KEY")
	default:
		return nil, config.ErrorWrongMessage
	}
	var reqToken string
	err = json.Unmarshal(reqBody, &reqToken)
	if err != nil {
		return nil, config.ErrorUnmarshal
	}
	valid, payload := mvt.VerifyMVTToken(headerSignature, reqToken, pubKey, os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return nil, config.ErrorInvalidPassword
	}
	return payload, nil
}
