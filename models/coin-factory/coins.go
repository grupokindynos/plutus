package coinfactory

import (
	"os"
	"strings"
)

type Coin struct {
	Tag     string
	RpcUser string
	RpcPass string
	RpcPort string
	Host    string
	Port    string
	User    string
	PrivKey string
}

// GetCoin is the safe way to check if a coin exists and retrieve the coin data
func GetCoin(tag string) *Coin {
	return &Coin{
		Tag:     strings.ToUpper(tag),
		RpcUser: os.Getenv(strings.ToUpper(tag) + "_RPC_USER"),
		RpcPass: os.Getenv(strings.ToUpper(tag) + "_RPC_PASS"),
		RpcPort: os.Getenv(strings.ToUpper(tag) + "_RPC_PORT"),
		Host:    os.Getenv(strings.ToUpper(tag) + "_IP"),
		Port:    os.Getenv(strings.ToUpper(tag) + "_SSH_PORT"),
		User:    os.Getenv(strings.ToUpper(tag) + "_SSH_USER"),
		PrivKey: os.Getenv(strings.ToUpper(tag) + "_SSH_PRIVKEY"),
	}
}
