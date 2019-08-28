package coinfactory

import (
	"github.com/grupokindynos/plutus/config"
	"os"
	"strings"
)

var polis = Coin{
	Tag:            "POLIS",
	ExternalSource: "blockbook.polispay.org",
	RpcMethods: RPCMethods{
		GetWalletInfo:     "getwalletinfo",
		GetBlockchainInfo: "getblockchaininfo",
		GetNetworkInfo:    "getnetworkinfo",
		GetNewAddress:     "getnewaddress",
		SendToAddress:     "sendtoaddress",
	},
}
var digibyte = Coin{
	Tag:            "DGB",
	ExternalSource: "dgb2.trezor.io",
	RpcMethods: RPCMethods{
		GetWalletInfo:     "getwalletinfo",
		GetBlockchainInfo: "getblockchaininfo",
		GetNetworkInfo:    "getnetworkinfo",
		GetNewAddress:     "getnewaddress",
		SendToAddress:     "sendtoaddress",
	},
}
var zcoin = Coin{
	Tag:            "XZC",
	ExternalSource: "xzc.polispay.com",
	RpcMethods: RPCMethods{
		GetWalletInfo:     "getwalletinfo",
		GetBlockchainInfo: "getblockchaininfo",
		GetNetworkInfo:    "getnetworkinfo",
		GetNewAddress:     "getnewaddress",
		SendToAddress:     "sendtoaddress",
	},
}
var litecoin = Coin{
	Tag:            "LTC",
	ExternalSource: "ltc2.trezor.io",
	RpcMethods: RPCMethods{
		GetWalletInfo:     "getwalletinfo",
		GetBlockchainInfo: "getblockchaininfo",
		GetNetworkInfo:    "getnetworkinfo",
		GetNewAddress:     "getnewaddress",
		SendToAddress:     "sendtoaddress",
	},
}
var bitcoin = Coin{
	Tag:            "BTC",
	ExternalSource: "btc2.trezor.io",
	RpcMethods: RPCMethods{
		GetWalletInfo:     "getwalletinfo",
		GetBlockchainInfo: "getblockchaininfo",
		GetNetworkInfo:    "getnetworkinfo",
		GetNewAddress:     "getnewaddress",
		SendToAddress:     "sendtoaddress",
	},
}
var dash = Coin{
	Tag:            "DASH",
	ExternalSource: "dash2.trezor.io",
	RpcMethods: RPCMethods{
		GetWalletInfo:     "getwalletinfo",
		GetBlockchainInfo: "getblockchaininfo",
		GetNetworkInfo:    "getnetworkinfo",
		GetNewAddress:     "getnewaddress",
		SendToAddress:     "sendtoaddress",
	},
}
var groestlcoin = Coin{
	Tag:            "GRS",
	ExternalSource: "grs.polispay.com",
	RpcMethods: RPCMethods{
		GetWalletInfo:     "getwalletinfo",
		GetBlockchainInfo: "getblockchaininfo",
		GetNetworkInfo:    "getnetworkinfo",
		GetNewAddress:     "getnewaddress",
		SendToAddress:     "sendtoaddress",
	},
}
var colossus = Coin{
	Tag:            "COLX",
	ExternalSource: "",
	RpcMethods: RPCMethods{
		GetWalletInfo:     "getwalletinfo",
		GetBlockchainInfo: "getblockchaininfo",
		GetNetworkInfo:    "getnetworkinfo",
		GetNewAddress:     "getnewaddress",
		SendToAddress:     "sendtoaddress",
	},
}
var deeponion = Coin{
	Tag:            "ONION",
	ExternalSource: "",
	RpcMethods: RPCMethods{
		GetWalletInfo:     "getwalletinfo",
		GetBlockchainInfo: "getblockchaininfo",
		GetNetworkInfo:    "getnetworkinfo",
		GetNewAddress:     "getnewaddress",
		SendToAddress:     "sendtoaddress",
	},
}
var mnpcoin = Coin{
	Tag:            "MNP",
	ExternalSource: "",
	RpcMethods: RPCMethods{
		GetWalletInfo:     "getwalletinfo",
		GetBlockchainInfo: "getblockchaininfo",
		GetNetworkInfo:    "getnetworkinfo",
		GetNewAddress:     "getnewaddress",
		SendToAddress:     "sendtoaddress",
	},
}

type RPCMethods struct {
	GetWalletInfo     string
	GetBlockchainInfo string
	GetNetworkInfo    string
	GetNewAddress     string
	SendToAddress     string
}

type Coin struct {
	ExternalSource  string
	RpcMethods      RPCMethods
	ColdAddress     string
	ExchangeAddress string
	Tag             string
	RpcUser         string
	RpcPass         string
	RpcPort         string
	Host            string
	Port            string
	User            string
	PrivKey         string
}

var Coins = map[string]*Coin{
	"POLIS": &polis,
	"DGB":   &digibyte,
	"XZC":   &zcoin,
	"LTC":   &litecoin,
	"BTC":   &bitcoin,
	"DASH":  &dash,
	"GRS":   &groestlcoin,
	"COLX":  &colossus,
	"ONION": &deeponion,
	"MNP":   &mnpcoin,
}

// GetCoin is the safe way to check if a coin exists and retrieve the coin data
func GetCoin(tag string) (*Coin, error) {
	coin, ok := Coins[strings.ToUpper(tag)]
	if !ok {
		return nil, config.ErrorNoCoin
	}
	coin = &Coin{
		Tag:             coin.Tag,
		ExternalSource:  coin.ExternalSource,
		RpcMethods:      coin.RpcMethods,
		ColdAddress:     os.Getenv(strings.ToUpper(tag) + "_RPC_USER"),
		RpcUser:         os.Getenv(strings.ToUpper(tag) + "_COLD_ADDRESS"),
		ExchangeAddress: os.Getenv(strings.ToUpper(tag) + "_EXCHANGE_ADDRESS"),
		RpcPass:         os.Getenv(strings.ToUpper(tag) + "_RPC_PASS"),
		RpcPort:         os.Getenv(strings.ToUpper(tag) + "_RPC_PORT"),
		Host:            os.Getenv(strings.ToUpper(tag) + "_IP"),
		Port:            os.Getenv(strings.ToUpper(tag) + "_SSH_PORT"),
		User:            os.Getenv(strings.ToUpper(tag) + "_SSH_USER"),
		PrivKey:         os.Getenv(strings.ToUpper(tag) + "_SSH_PRIVKEY"),
	}
	return coin, nil
}
