package coinfactory

import (
	"github.com/grupokindynos/plutus/config"
	"os"
	"strings"
)

type RPCMethods struct {
	GetWalletInfo     string
	GetBlockchainInfo string
	GetNetworkInfo    string
	GetNewAddress     string
}

type Coin struct {
	Tag        string
	RpcUser    string
	RpcPass    string
	RpcPort    string
	Host       string
	Port       string
	User       string
	PrivKey    string
	RpcMethods RPCMethods
}

type Coins []Coin

func GetRPCMethods(coin *Coin) RPCMethods {
	if coin.Tag == "ETH" {
		methods := RPCMethods{
			GetWalletInfo:     "",
			GetBlockchainInfo: "",
			GetNetworkInfo:    "",
			GetNewAddress:     "personal_newAccount",
		}
		return methods
	} else {
		methods := RPCMethods{
			GetWalletInfo:     "getwalletinfo",
			GetBlockchainInfo: "getblockchaininfo",
			GetNetworkInfo:    "getnetworkinfo",
			GetNewAddress:     "getnewaddress",
		}
		return methods
	}
}

// GetCoin is the safe way to check if a coin exists and retrieve the coin data
func GetCoin(tag string) (*Coin, error) {
	host := os.Getenv(strings.ToUpper(tag) + "_IP")
	if host == "" {
		return nil, config.ErrorNoCoin
	}
	coin := &Coin{
		Tag:     strings.ToUpper(tag),
		RpcUser: os.Getenv(strings.ToUpper(tag) + "_RPC_USER"),
		RpcPass: os.Getenv(strings.ToUpper(tag) + "_RPC_PASS"),
		RpcPort: os.Getenv(strings.ToUpper(tag) + "_RPC_PORT"),
		Host:    os.Getenv(strings.ToUpper(tag) + "_IP"),
		Port:    os.Getenv(strings.ToUpper(tag) + "_SSH_PORT"),
		User:    os.Getenv(strings.ToUpper(tag) + "_SSH_USER"),
		PrivKey: os.Getenv(strings.ToUpper(tag) + "_SSH_PRIVKEY"),
	}
	coin.RpcMethods = GetRPCMethods(coin)
	return coin, nil
}
