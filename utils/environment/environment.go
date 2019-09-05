package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/grupokindynos/plutus/config"
	coinfactory "github.com/grupokindynos/plutus/models/coin-factory"
	heroku "github.com/heroku/heroku-go/v5"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sethvargo/go-password/password"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type EnvironmentVars struct {
	HerokuUsername string
	HerokuPassword string
	AuthUsername   string
	AuthPassword   string
	GinMode        string
	KeyPassword    string
	CoinsVars      []CoinVar
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

var (
	NewVars    EnvironmentVars
	OldVars    EnvironmentVars
	HerokuUser string
	HerokuPass string
)

// This script will only work with a full set of environment variables.
// Should only be used to recreate ssh keys and passwords
func main() {
	// First load the current .env
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	HerokuUser = os.Getenv("HEROKU_USERNAME")
	HerokuPass = os.Getenv("HEROKU_PASSWORD")
	if HerokuUser == "" || HerokuPass == "" {
		panic(errors.New("no heroku login details, we can't continue"))
	}
	heroku.DefaultTransport.Username = HerokuUser
	heroku.DefaultTransport.Password = HerokuPass
	h := heroku.NewService(heroku.DefaultClient)
	_, err = h.AppInfo(context.Background(), "plutus-wallets")
	if err != nil {
		panic(err)
	}
	log.Println("Creating environment file...")
	// Move current .env file to a backup file
	date := time.Now().Format("2006-01-02")
	err = os.Rename("../../.env", "../../old-env-backup-"+date)
	if err != nil {
		log.Fatal("Error moving .env file")
	}
	// Get and object with old variables
	OldVars = getOldVars()
	// Generate new keys
	NewVars = genNewVars()
	log.Println("Updating remote keys...")
	// Update new ssh keys with a ssh client using old keys
	for _, server := range OldVars.CoinsVars {
		var newCoinPubKey string
		for _, newCoinVar := range NewVars.CoinsVars {
			if newCoinVar.Coin == server.Coin {
				newCoinPubKey = newCoinVar.SshPubKey
			}
		}
		log.Println("Updating remote for " + server.Coin)
		err := updateRemoteKey(server, newCoinPubKey)
		if err != nil {
			panic(err)
		}
	}
	// Update heroku environment variables.

	// Create environment map
	envMap := map[string]*string{
		"AUTH_PASSWORD": &NewVars.AuthPassword,
		"AUTH_USERNAME": &NewVars.AuthUsername,
		"KEY_PASSWORD":  &NewVars.KeyPassword,
		"GIN_MODE":      &NewVars.GinMode,
	}
	for _, env := range NewVars.CoinsVars {
		envMap[strings.ToUpper(env.Coin)+"_IP"] = &env.SshHost
		envMap[strings.ToUpper(env.Coin)+"_RPC_USER"] = &env.RpcUser
		envMap[strings.ToUpper(env.Coin)+"_RPC_PASS"] = &env.RpcPass
		envMap[strings.ToUpper(env.Coin)+"_RPC_PORT"] = &env.RpcPort
		envMap[strings.ToUpper(env.Coin)+"_SSH_USER"] = &env.SshUser
		envMap[strings.ToUpper(env.Coin)+"_SSH_PORT"] = &env.SshPort
		envMap[strings.ToUpper(env.Coin)+"_SSH_PRIVKEY"] = &env.SshPrivKey
		envMap[strings.ToUpper(env.Coin)+"_COLD_ADDRESS"] = &env.ColdAddrs
		envMap[strings.ToUpper(env.Coin)+"_EXCHANGE_ADDRESS"] = &env.ExchangeAddrs
	}
	res, err := h.ConfigVarUpdate(context.Background(), "plutus-wallets", envMap)
	if err != nil {
		panic("critical error, unable to update heroku variables, need manual correction urgently!")
	}
	fmt.Println(res)
	// Dump new keys to .env file
	err = saveNewVars()
	if err != nil {
		panic(err)
	}
}

func updateRemoteKey(coinVars CoinVar, newCoinPubKey string) error {
	privKey, err := parsePrivKey(coinVars.SshPrivKey, OldVars.KeyPassword)
	if err != nil {
		return err
	}
	sshConfig := &ssh.ClientConfig{
		User: coinVars.SshUser,
		Auth: []ssh.AuthMethod{privKey},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	connection, err := ssh.Dial("tcp", coinVars.SshHost+":"+coinVars.SshPort, sshConfig)
	if err != nil {
		fmt.Println(err)
		return err
	}
	session, err := connection.NewSession()
	if err != nil {
		fmt.Println(err)
		return err
	}
	// First cmd remove the second line (the first line is the main key and must be changed manually to prevent removing all access to server)
	// Second cmd append the newCoinPubKey to the authorized_keys file
	cmd := "sed -i '2d' .ssh/authorized_keys && sed -i '$a" + newCoinPubKey + "' .ssh/authorized_keys"
	err = session.Run(cmd)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func getOldVars() EnvironmentVars {
	Vars := EnvironmentVars{
		HerokuUsername: os.Getenv("HEROKU_USERNAME"),
		HerokuPassword: os.Getenv("HEROKU_PASSWORD"),
		AuthUsername:   os.Getenv("AUTH_USERNAME"),
		AuthPassword:   os.Getenv("AUTH_PASSWORD"),
		GinMode:        os.Getenv("GIN_MODE"),
		KeyPassword:    os.Getenv("KEY_PASSWORD"),
		CoinsVars:      nil,
	}
	for key := range coinfactory.Coins {
		coinVars := CoinVar{
			Coin:          strings.ToUpper(key),
			RpcUser:       os.Getenv(strings.ToUpper(key) + "_RPC_USER"),
			RpcPass:       os.Getenv(strings.ToUpper(key) + "_RPC_PASS"),
			RpcPort:       os.Getenv(strings.ToUpper(key) + "_RPC_PORT"),
			SshUser:       os.Getenv(strings.ToUpper(key) + "_SSH_USER"),
			SshPrivKey:    os.Getenv(strings.ToUpper(key) + "_SSH_PRIVKEY"),
			SshPubKey:     "",
			SshPort:       os.Getenv(strings.ToUpper(key) + "_SSH_PORT"),
			SshHost:       os.Getenv(strings.ToUpper(key) + "_IP"),
			ExchangeAddrs: os.Getenv(strings.ToUpper(key) + "_EXCHANGE_ADDRESS"),
			ColdAddrs:     os.Getenv(strings.ToUpper(key) + "_COLD_ADDRESS"),
		}
		Vars.CoinsVars = append(Vars.CoinsVars, coinVars)
	}
	return Vars
}

func genNewVars() EnvironmentVars {
	newAuthUsername := generateRandomPassword(128)
	newAuthPassword := generateRandomPassword(128)
	newDecryptionKey := generateRandomPassword(32)

	Vars := EnvironmentVars{
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
	return Vars
}

func saveNewVars() error {
	err := ioutil.WriteFile("../../.env", []byte(NewVars.ToString()), 777)
	if err != nil {
		return err
	}
	return nil
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

func parsePrivKey(privKey string, encryptionKey string) (ssh.AuthMethod, error) {
	decrypted, err := config.Decrypt([]byte(encryptionKey), privKey)
	if err != nil {
		return nil, err
	}
	key, err := ssh.ParsePrivateKey([]byte(decrypted))
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}
