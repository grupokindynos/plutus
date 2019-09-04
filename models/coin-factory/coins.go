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
		GetWalletInfo:              "getwalletinfo",
		GetBlockchainInfo:          "getblockchaininfo",
		GetNetworkInfo:             "getnetworkinfo",
		GetNewAddress:              "getnewaddress",
		SendToAddress:              "sendtoaddress",
		ValidateAddress:            "validateaddress",
		GetRawTransaction:          "getrawtransaction",
		GetRawTransactionVerbosity: true,
	},
}
var digibyte = Coin{
	Tag:            "DGB",
	ExternalSource: "dgb2.trezor.io",
	RpcMethods: RPCMethods{
		GetWalletInfo:              "getwalletinfo",
		GetBlockchainInfo:          "getblockchaininfo",
		GetNetworkInfo:             "getnetworkinfo",
		GetNewAddress:              "getnewaddress",
		SendToAddress:              "sendtoaddress",
		ValidateAddress:            "getaddressinfo",
		GetRawTransaction:          "getrawtransaction",
		GetRawTransactionVerbosity: "1",
	},
}
var zcoin = Coin{
	Tag:            "XZC",
	ExternalSource: "xzc.polispay.com",
	RpcMethods: RPCMethods{
		GetWalletInfo:              "getwalletinfo",
		GetBlockchainInfo:          "getblockchaininfo",
		GetNetworkInfo:             "getnetworkinfo",
		GetNewAddress:              "getnewaddress",
		SendToAddress:              "sendtoaddress",
		ValidateAddress:            "validateaddress",
		GetRawTransaction:          "getrawtransaction",
		GetRawTransactionVerbosity: true,
	},
}
var litecoin = Coin{
	Tag:            "LTC",
	ExternalSource: "ltc2.trezor.io",
	RpcMethods: RPCMethods{
		GetWalletInfo:              "getwalletinfo",
		GetBlockchainInfo:          "getblockchaininfo",
		GetNetworkInfo:             "getnetworkinfo",
		GetNewAddress:              "getnewaddress",
		SendToAddress:              "sendtoaddress",
		ValidateAddress:            "getaddressinfo",
		GetRawTransaction:          "getrawtransaction",
		GetRawTransactionVerbosity: "1",
	},
}
var bitcoin = Coin{
	Tag:            "BTC",
	ExternalSource: "btc2.trezor.io",
	RpcMethods: RPCMethods{
		GetWalletInfo:              "getwalletinfo",
		GetBlockchainInfo:          "getblockchaininfo",
		GetNetworkInfo:             "getnetworkinfo",
		GetNewAddress:              "getnewaddress",
		SendToAddress:              "sendtoaddress",
		ValidateAddress:            "getaddressinfo",
		GetRawTransaction:          "getrawtransaction",
		GetRawTransactionVerbosity: "1",
	},
}
var dash = Coin{
	Tag:            "DASH",
	ExternalSource: "dash2.trezor.io",
	RpcMethods: RPCMethods{
		GetWalletInfo:              "getwalletinfo",
		GetBlockchainInfo:          "getblockchaininfo",
		GetNetworkInfo:             "getnetworkinfo",
		GetNewAddress:              "getnewaddress",
		SendToAddress:              "sendtoaddress",
		ValidateAddress:            "validateaddress",
		GetRawTransaction:          "getrawtransaction",
		GetRawTransactionVerbosity: true,
	},
}
var groestlcoin = Coin{
	Tag:            "GRS",
	ExternalSource: "grs.polispay.com",
	RpcMethods: RPCMethods{
		GetWalletInfo:              "getwalletinfo",
		GetBlockchainInfo:          "getblockchaininfo",
		GetNetworkInfo:             "getnetworkinfo",
		GetNewAddress:              "getnewaddress",
		SendToAddress:              "sendtoaddress",
		ValidateAddress:            "getaddressinfo",
		GetRawTransaction:          "getrawtransaction",
		GetRawTransactionVerbosity: "1",
	},
}
var colossus = Coin{
	Tag:            "COLX",
	ExternalSource: "",
	RpcMethods: RPCMethods{
		GetWalletInfo:              "getwalletinfo",
		GetBlockchainInfo:          "getblockchaininfo",
		GetNetworkInfo:             "getnetworkinfo",
		GetNewAddress:              "getnewaddress",
		SendToAddress:              "sendtoaddress",
		ValidateAddress:            "validateaddress",
		GetRawTransaction:          "getrawtransaction",
		GetRawTransactionVerbosity: true,
	},
}
var deeponion = Coin{
	Tag:            "ONION",
	ExternalSource: "",
	RpcMethods: RPCMethods{
		GetWalletInfo:              "getwalletinfo",
		GetBlockchainInfo:          "getblockchaininfo",
		GetNetworkInfo:             "getnetworkinfo",
		GetNewAddress:              "getnewaddress",
		SendToAddress:              "sendtoaddress",
		ValidateAddress:            "validateaddress",
		GetRawTransaction:          "getrawtransaction",
		GetRawTransactionVerbosity: true,
	},
}
var mnpcoin = Coin{
	Tag:            "MNP",
	ExternalSource: "",
	RpcMethods: RPCMethods{
		GetWalletInfo:              "getwalletinfo",
		GetBlockchainInfo:          "getblockchaininfo",
		GetNetworkInfo:             "getnetworkinfo",
		GetNewAddress:              "getnewaddress",
		SendToAddress:              "sendtoaddress",
		ValidateAddress:            "validateaddress",
		GetRawTransaction:          "getrawtransaction",
		GetRawTransactionVerbosity: true,
	},
}
var snowgem = Coin{
	Tag:            "XSG",
	ExternalSource: "",
	RpcMethods: RPCMethods{
		GetWalletInfo:              "getwalletinfo",
		GetBlockchainInfo:          "getblockchaininfo",
		GetNetworkInfo:             "getnetworkinfo",
		GetNewAddress:              "getnewaddress",
		SendToAddress:              "sendtoaddress",
		ValidateAddress:            "validateaddress",
		GetRawTransaction:          "getrawtransaction",
		GetRawTransactionVerbosity: true,
	},
}
var ethereum = Coin{
	Tag:            "ETH",
	ExternalSource: "",
	RpcMethods: RPCMethods{
		GetWalletInfo:     "",
		GetBlockchainInfo: "",
		GetNetworkInfo:    "",
		GetNewAddress:     "",
		SendToAddress:     "",
	},
}

type RPCMethods struct {
	GetWalletInfo              string
	GetBlockchainInfo          string
	GetNetworkInfo             string
	GetNewAddress              string
	SendToAddress              string
	ValidateAddress            string
	GetRawTransaction          string
	GetRawTransactionVerbosity interface{}
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
	"XSG":   &snowgem,
	"ETH":   &ethereum,
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
		ColdAddress:     os.Getenv(strings.ToUpper(tag) + "_COLD_ADDRESS"),
		ExchangeAddress: os.Getenv(strings.ToUpper(tag) + "_EXCHANGE_ADDRESS"),
		RpcUser:         os.Getenv(strings.ToUpper(tag) + "_RPC_USER"),
		RpcPass:         os.Getenv(strings.ToUpper(tag) + "_RPC_PASS"),
		RpcPort:         os.Getenv(strings.ToUpper(tag) + "_RPC_PORT"),
		Host:            os.Getenv(strings.ToUpper(tag) + "_IP"),
		Port:            os.Getenv(strings.ToUpper(tag) + "_SSH_PORT"),
		User:            os.Getenv(strings.ToUpper(tag) + "_SSH_USER"),
		PrivKey:         os.Getenv(strings.ToUpper(tag) + "_SSH_PRIVKEY"),
	}
	return coin, nil
}

func CheckCoinConfigs(coin *Coin) error {
	if coin.Tag != "ETH" {
		if coin.RpcUser == "" {
			return config.ErrorNoRpcUserProvided
		}
		if coin.RpcPass == "" {
			return config.ErrorNoRpcPassProvided
		}
	}
	if coin.RpcPort == "" {
		return config.ErrorNoRpcPortProvided
	}
	if coin.Host == "" {
		return config.ErrorNoHostIPProvided
	}
	if coin.Port == "" {
		return config.ErrorNoHostPortProvided
	}
	if coin.User == "" {
		return config.ErrorNoHostUserProvided
	}
	if coin.PrivKey == "" {
		return config.ErrorNoAuthMethodProvided
	}
	if coin.ExchangeAddress == "" {
		return config.ErrorNoExchangeAddress
	}
	if coin.ColdAddress == "" {
		return config.ErrorNoColdAddress
	}

	return nil
}
