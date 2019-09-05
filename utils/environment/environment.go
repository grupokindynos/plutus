package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/grupokindynos/plutus/config"
	coinfactory "github.com/grupokindynos/plutus/models/coin-factory"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sethvargo/go-password/password"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type EnvironmentVars struct {
	AuthUsername string
	AuthPassword string
	GinMode      string
	KeyPassword  string
	CoinsVars    []CoinVar
}

func (ev *EnvironmentVars) ToString() string {
	str := "" +
		"AUTH_USERNAME=" + ev.AuthUsername + "\n" +
		"AUTH_PASSWORD=" + ev.AuthPassword + "\n" +
		"KEY_PASSWORD=" + ev.KeyPassword + "\n" +
		"GIN_MODE=" + ev.GinMode + "\n"
	for _, coinVar := range ev.CoinsVars {
		str += coinVar.ToString()
	}
	return str
}

type CoinVar struct {
	Coin          string
	RpcUser       string
	RpcPass       string
	RpcPort       string
	SshUser       string
	SshHost       string
	SshPrivKey    string
	SshPubKey     string
	SshPort       string
	ExchangeAddrs string
	ColdAddrs     string
}

func (cv *CoinVar) ToString() string {
	str := "" +
		strings.ToUpper(cv.Coin) + "_RPC_USER=" + cv.RpcUser + "\n" +
		strings.ToUpper(cv.Coin) + "_RPC_PASS=" + cv.RpcPass + "\n" +
		strings.ToUpper(cv.Coin) + "_RPC_PORT=" + cv.RpcPort + "\n" +
		strings.ToUpper(cv.Coin) + "_SSH_USER=" + cv.SshUser + "\n" +
		strings.ToUpper(cv.Coin) + "_IP=" + cv.SshHost + "\n" +
		strings.ToUpper(cv.Coin) + "_SSH_PRIVKEY=" + cv.SshPrivKey + "\n" +
		strings.ToUpper(cv.Coin) + "_SSH_PORT=" + cv.SshPort + "\n" +
		strings.ToUpper(cv.Coin) + "_EXCHANGE_ADDRESS=" + cv.ExchangeAddrs + "\n" +
		strings.ToUpper(cv.Coin) + "_COLD_ADDRESS=" + cv.ColdAddrs + "\n"
	return str
}

type KeyPair struct {
	Private []byte
	Public  []byte
}

var Vars EnvironmentVars

// This script will only work with a full set of environment variables.
// Should only be used to recreate ssh keys and passwords
func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println("Creating environment file")
	date := time.Now().Format("2006-01-02")
	err = os.Rename("../../.env", "../../old-env-backup-" + date)
	if err != nil {
		log.Fatal("Error moving .env file")
	}
	newAuthUsername := generateRandomPassword(128)
	newAuthPassword := generateRandomPassword(128)
	newDecryptionKey := generateRandomPassword(32)

	Vars = EnvironmentVars{
		AuthUsername: newAuthUsername,
		AuthPassword: newAuthPassword,
		GinMode:      os.Getenv("GIN_MODE"),
		KeyPassword:  newDecryptionKey,
		CoinsVars:    nil,
	}
	for key := range coinfactory.Coins {
		log.Println("Creating vars for " + strings.ToUpper(key))
		keyPair := genPrivKeyPair()
		encryptedPrivKey, err := config.Encrypt([]byte(newDecryptionKey), keyPair.Private)
		if err != nil {
			panic(err)
		}
		coinVars := CoinVar{
			Coin:          strings.ToUpper(key),
			RpcUser:       os.Getenv(strings.ToUpper(key) + "_RPC_USER"),
			RpcPass:       os.Getenv(strings.ToUpper(key) + "_RPC_PASS"),
			RpcPort:       os.Getenv(strings.ToUpper(key) + "_RPC_PORT"),
			SshUser:       os.Getenv(strings.ToUpper(key) + "_SSH_USER"),
			SshPrivKey:    encryptedPrivKey,
			SshPubKey:     string(keyPair.Public),
			SshPort:       os.Getenv(strings.ToUpper(key) + "_SSH_PORT"),
			SshHost:       os.Getenv(strings.ToUpper(key) + "_IP"),
			ExchangeAddrs: os.Getenv(strings.ToUpper(key) + "_EXCHANGE_ADDRESS"),
			ColdAddrs:     os.Getenv(strings.ToUpper(key) + "_COLD_ADDRESS"),
		}
		Vars.CoinsVars = append(Vars.CoinsVars, coinVars)
	}
	err = ioutil.WriteFile("../../.env", []byte(Vars.ToString()), 777)
	if err != nil {
		panic(err)
	}
}

func genPrivKeyPair() KeyPair {
	bitSize := 4096
	privateKey, err := generatePrivateKey(bitSize)
	if err != nil {
		log.Fatal(err.Error())
	}
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}
	privateBytes := pem.EncodeToMemory(&privBlock)
	publicKeyBytes, err := generatePublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	return KeyPair{Private: privateBytes, Public: publicKeyBytes}
}

func generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func generatePublicKey(privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}
	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)
	return pubKeyBytes, nil
}

func generateRandomPassword(size int) string {
	res, err := password.Generate(size, 10, 0, false, true)
	if err != nil {
		log.Fatal(err)
	}
	return res
}
