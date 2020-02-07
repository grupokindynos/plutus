package controllers

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/eabz/btcutil"
	"github.com/eabz/btcutil/chaincfg"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/grupokindynos/common/blockbook"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/plutus"
	"github.com/martinboehm/btcd/wire"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"testing"
)

func init() {
	var i uint32
	for key, _ := range coinfactory.Coins {
		coinConf, err := coinfactory.GetCoin(key)
		if err != nil {
			panic(err)
		}
		if !coinConf.Info.Token && coinConf.Info.Tag != "ETH" {
			i += 1
			coinConf.NetParams.AddressMagicLen = 1
			coinConf.NetParams.Net = wire.BitcoinNet(i)
			err := chaincfg.Register(coinConf.NetParams)
			if err != nil {
				panic(err)
			}
		}
	}

}

var testMnemonic = "maximum potato bitter govern rebuild elegant nest boring note caution wedding exercise near chimney narrow"

type testData struct {
	coin    *coins.Coin
	xpub    string
	xprv    string
	path    uint32
	privKey string
	addr    string
}

var testXpup = []testData{
	{path: 10, addr: "1JsMio1jkCBsneBsQxs1cbpj5BwTnZKS8i", privKey: "L24bVGEqUBnQ1pUxoWGtBordNveobDcdxU3zDvkoVWK5RxrYeTde", coin: coinfactory.Coins["BTC"], xprv: "xprv9yK7buzAoK2xTZz5WU19MYm1NqyCbfLZ8JVz9yPLc7BDQ8mjMMr3uexVfUzYveaaELMq2cAXEw4ErRDQZ7JhTyq1DQKimoym18LQ2WX6noa", xpub: "xpub6CJU1RX4dgbFg44YcVY9ighjvsoh184QVXRaxMnxASiCGw6stuAJTTGyWkvyv7d2HKMz2V9hUFBWfYQCjFZUDrxna82vURQTVwkp69poMhx"},
	{path: 10, addr: "XovGmrsTDduAuUdaAMmTCDJ3yx1c8i7bt5", privKey: "XKfbqtXi1mat3YcJ6imRT4x2mWZ6Majo4AzUBgP8iW9VSF1zdyXb", coin: coinfactory.Coins["DASH"], xprv: "xprv9zUu1AYZf7HxbewYJSrasPNhoTj6DTpNgxSHcjWkfNMpwj3WQJYqSTVC2uK96KKtfukA4jLoTqQBqAohTjAeTmmYKzP3VzUxVZxBgmNXYhC", xpub: "xpub6DUFQg5TVUrFp921QUPbEXKSMVZacvYE4BMtR7vNDhtopXNewqs5zFoftATvsgAexA87pgmPCd2hNXfhWcR9UbkuFkShFWgkPvtzgmJragb"},
	{path: 10, addr: "D8tmU27w7vAyw37KHZbZz1Htsefz9RMoFz", privKey: "L1xLSnxKojULWLV45xU4Wtvu7yv6EQ7CWc55z6JtbXN4Sx3Jw4i5", coin: coinfactory.Coins["DGB"], xprv: "xprv9y4A8Z4j49fBpUnr6tQ2A2tEJBFZkk4HEAaLRbfret9QpCJ3NFfhpWem9WDbaDpic1jGcfd3QuNyDPVsV14xTCH767ZiL4dUyvbJChK82FU", xpub: "xpub6C3WY4bctXDV2xsKCuw2XApxrD64ACn8bPVwDz5UDDgPgzdBunyxNJyEzp7Qt5sf3YEC8UHezceZt2uixmyxsgx5y6CVvxb4eBwJhyCNXZC"},
	{path: 10, addr: "FpZhQAurnQM5LUZy9KxUmAfzY9P8qqC696", privKey: "L26v51gauTKN6iGDSD6QgCVXxmnA3YvyHByPGp3YejhJq42fJipc", coin: coinfactory.Coins["GRS"], xprv: "xprv9yhswrkeWDmwhdz4bvSJJbRzpEtBQcLyMarWrptp6hGSRg2gbq4oE9e5pFa8RRcHw2bT4tzydjnhhaf7HGcBcLTfKANm5qDmRBSzREGPH5D", xpub: "xpub6ChEMNHYLbLEv84XhwyJfjNjNGifp54pion7fDJRf2oRJUMq9NP3mwxZfXDKYZaEWN5MYBE1PFmbNfUBNgJiezXwWjnh6KAW2LYf5U5AbvH"},
	{path: 10, addr: "a2LxezdJMB86HA4WgHJM2WaAcapEPcnSGn", privKey: "Y8izsjrgrwYh6jaXWn9tqUn4wc66GntRTw7mta2z5HNY12QL8xqu", coin: coinfactory.Coins["XZC"], xprv: "xprv9yKC8QCqAaQBG71sEdhXcDWzko1K6LkbeqMRptVefrMkF5yjz8z6BkqXMp514NSz1NnEajk66NybkKnryzPiSDzLhqRmM8RarQhA17m7a7d", xpub: "xpub6CJYXujizwxUUb6LLfEXyMTjJpqoVoUT24H2dGuGEBtj7tJtXgJLjZA1D3u7TAnYtAQd1AuAfF8UXtQhW6YxwwQ6dDe8MEdTkLPYGSRJ2B2"},
	{path: 10, addr: "LhGukzG7eKsR2MRDUCLWVgqEGsetnM55QK", privKey: "TAxdgg1MXmUdvhoUpX5NA3deShB4z9PVrhbBQxNVDhey5Keh1WHk", coin: coinfactory.Coins["LTC"], xprv: "xprv9yXzgYyupVCWwCk5eR7kd9oV3x5PDjZ9HLT5hjRcYsGbAsy2f6dCJugJWX94LTtMhYSbiRfynmj37W4aVbrQGhX9ekMcufPY6AY8f1KkG8V", xpub: "xpub6CXM64Woerkp9gpYkSekzHkDbyusdCGzeZNgW7qE7Coa3gJBCdwSrhznMmpH2RyVFwyDxTMBQexPFMMg7J3ujZzoMFLtiK6aVMWVxAUtAA7"},
}

func TestXpubGeneration(t *testing.T) {
	for _, test := range testXpup {
		test.coin.Mnemonic = testMnemonic
		acc, err := getAccFromMnemonic(test.coin, true)
		if err != nil {
			panic(err)
		}
		equalPriv := reflect.DeepEqual(test.xprv, acc.String())
		if !equalPriv {
			t.Error("xpriv doesn't match for " + test.coin.Info.Tag + " expected: " + test.xprv + " got: " + acc.String())
		}
		accPub, err := acc.Neuter()
		if err != nil {
			panic(err)
		}
		equalXpub := reflect.DeepEqual(test.xpub, accPub.String())
		if !equalXpub {
			t.Error("xpub doesn't match for " + test.coin.Info.Tag + " expected: " + test.xpub + " got: " + accPub.String())
		}
		pubKeyHash, err := getPubKeyHashFromPath(acc, test.coin, test.path)
		if err != nil {
			panic(err)
		}
		equalAddr := reflect.DeepEqual(test.addr, pubKeyHash)
		if !equalAddr {
			t.Error("addr doesn't match for " + test.coin.Info.Tag + " expected: " + test.addr + " got: " + pubKeyHash)
		}
		privKey, err := getPrivKeyFromPath(acc, test.path)
		wif, err := btcutil.NewWIF(privKey, test.coin.NetParams, true, test.coin.NetParams.Base58CksumHasher)
		if err != nil {
			panic(err)
		}
		equalPrivKey := reflect.DeepEqual(test.privKey, wif.String())
		if !equalPrivKey {
			t.Error("privKey doesn't match for " + test.coin.Info.Tag + " expected: " + test.privKey + " got: " + wif.String())
		}
	}
}
func TestEthBalance(t *testing.T) {
	var acc3 = "0x931D387731bBbC988B312206c74F77D004D6B84b"
	//var acc2 = "0x71c7656ec7ab88b098defb751b7401b5f6d8976f"
	coinConfig, err := coinfactory.GetCoin("ETH")
	ethConfig, err := coinfactory.GetCoin("ETH")
	if err != nil {
		fmt.Println(err)
		//return nil, err
	}
	blockBookWrap := blockbook.NewBlockBookWrapper(ethConfig.Info.Blockbook)
	info, err := blockBookWrap.GetEthAddress(acc3)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	if coinConfig.Info.Token {
		var tokenInfo *blockbook.EthTokens
		for _, token := range info.Tokens {
			if coinConfig.Info.Contract == token.Contract {
				tokenInfo = &token
			}
		}
		if tokenInfo == nil {
			response := plutus.Balance{
				Confirmed: 0,
			}
			fmt.Println(response)
		} else {
			fmt.Println(tokenInfo)
			balance, err := strconv.ParseFloat(tokenInfo.Balance, 64)
			if err != nil {
				fmt.Println(err)
			}
			response := plutus.Balance{
				Confirmed: balance,
			}
			fmt.Println(response)
		}
	} else {
		balance, err := strconv.ParseFloat(info.Balance, 64)
		if err != nil {
			fmt.Println(err)
		}
		response := plutus.Balance{
			Confirmed: balance / 1e18,
		}
		fmt.Println(response)
	}
}

func TestERC20Balance(t *testing.T) {
	var acc3 = "0xC6BD3EDD07e294CB66B8318356D688b3516EA950"
	coinConfig, err := coinfactory.GetCoin("USDT")
	ethConfig, err := coinfactory.GetCoin("ETH")
	if err != nil {
		t.Error(err)
	}
	blockBookWrap := blockbook.NewBlockBookWrapper(ethConfig.Info.Blockbook)
	info, err := blockBookWrap.GetEthAddress(acc3)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(info)
	if coinConfig.Info.Token {
		var tokenInfo *blockbook.EthTokens
		for _, token := range info.Tokens {

			if coinConfig.Info.Contract == token.Contract {
				tokenInfo = &token
				break
			}
		}
		if tokenInfo == nil {
			response := plutus.Balance{
				Confirmed: 0,
			}
			fmt.Println(response)
			return
		}
		balance, err := strconv.ParseFloat(tokenInfo.Balance, 64)
		balance = balance / (math.Pow(10, float64(tokenInfo.Decimals)))
		if err != nil {
			t.Error(err)
		}
		response := plutus.Balance{
			Confirmed: balance,
		}
		fmt.Println(response)
		return
	} else {
		balance, err := strconv.ParseFloat(info.Balance, 64)
		if err != nil {
			t.Error(err)
		}
		response := plutus.Balance{
			Confirmed: balance / 1e18,
		}
		fmt.Println(response)
		return
	}
}

func TestValidateERC20raxTx(t *testing.T) {
	//rawTx := "f8aa1b85012a05f20083030d4094dac17f958d2ee523a2206206994597c13d831ec780b844a9059cbb000000000000000000000000673153460d01a22f9dac129f2ea59be3681921a400000000000000000000000000000000000000000000000000000000001e84801ba0a2437e58ab016fc47ae3a8b119bec36482549ca07d19eeebe373bff226a5823ba063134b2a2f4725b9879e484b59e22ebaf07083492528ecf0a43b7c94f8b8ddc5"
	rawTx := "f8a91b84b2d05e0083030d4094dac17f958d2ee523a2206206994597c13d831ec780b844a9059cbb000000000000000000000000673153460d01a22f9dac129f2ea59be3681921a400000000000000000000000000000000000000000000000000000000002f4d601ca0accf91f7628d757135bb2ec51afdd990a029e98c0ef19cf4673799501d241e76a051208e147977fb79c09e06885687637787efa8b34d46b398d98b5ee5da10c081"
	value := int64(3100000)
	var tx *types.Transaction
	rawtx, err := hex.DecodeString(rawTx)
	if err != nil {
		t.Error(err)
	}
	err = rlp.DecodeBytes(rawtx, &tx)
	if err != nil {
		t.Error(err)
	}
	var isValue bool
	var isAddress bool
	//compare amount from the tx and the input body
	bodyAddr := common.HexToAddress("0x673153460D01A22F9dAc129F2Ea59be3681921A4")
	//compare the address from the tx and the input body
	fmt.Println(hex.EncodeToString(tx.Data()))
	fmt.Println(tx.Data())
	address, amount := ReverseTransferPayload([]byte(hex.EncodeToString(tx.Data())))
	toAddr := common.HexToAddress(string(address))
	fmt.Println(toAddr.Hex())

	if amount.Int64() == value {
		isValue = true
	}
	if bytes.Equal(bodyAddr.Bytes(), toAddr.Bytes()) {
		isAddress = true
	}
	fmt.Println(isValue, isAddress)
}
func ReverseTransferPayload(b []byte) ([]byte, *big.Int) {
	to := b[32:72]
	tokens := b[74:136]
	hexed, _ := hex.DecodeString(string(tokens))
	amount := big.NewInt(0)
	amount.SetBytes(hexed)
	return to, amount
}
