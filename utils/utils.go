package utils

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/common/jwt"
	"github.com/grupokindynos/plutus/config"
	"os"
)

func VerifyHeaderSignature(c *gin.Context) (string, error) {
	// Get the microservice name from header
	verificationToken := c.GetHeader("service")
	if verificationToken == "" {
		return "", errors.New("missing header signature")
	}
	// Decode not-verify
	payload, err := jwt.DecodeJWSNoVerify(verificationToken)
	// Unmarshal the payload
	var serviceStr string
	err = json.Unmarshal(payload, &serviceStr)
	if err != nil {
		return "", err
	}
	pubKey, err := GetPubKeyFromStrService(serviceStr)
	// Verify the header token signature
	_, err = jwt.DecodeJWS(verificationToken, pubKey)
	// If there is an error, means the request was not properly signed
	if err != nil {
		return "", err
	}
	return pubKey, nil
}

func GetPubKeyFromStrService(service string) (pubKey string, err error) {
	switch service {
	case "ladon":
		pubKey = os.Getenv("LADON_PUBLIC_KEY")
	case "tyche":
		pubKey = os.Getenv("TYCHE_PUBLIC_KEY")
	case "adrestia":
		pubKey = os.Getenv("ADRESTIA_PUBLIC_KEY")
	default:
		return "", config.ErrorNoAuthorized
	}
	return pubKey, nil
}
