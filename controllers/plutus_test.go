package controllers

import (
	"encoding/hex"
	"fmt"
	"github.com/eabz/btcutil"
	"github.com/eabz/btcutil/chaincfg"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/grupokindynos/common/blockbook"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/plutus"
	"github.com/martinboehm/btcd/wire"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"golang.org/x/crypto/sha3"
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
	//acc, err := getEthAccFromMnemonic(coinConfig, false)
	//if err != nil {
	//	t.Error(err)
	//}
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

func TestERC20Transfer(t *testing.T) {
	mnem := "roof stable huge chuckle where else sniff apology museum maze parade delay"
	wallet, err := hdwallet.NewFromMnemonic(mnem)
	if err != nil {
		t.Error(err)
	}
	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, true)
	if err != nil {
		t.Error(err)
	}

	coinConfig, err := coinfactory.GetCoin("ETH")
	if err != nil {
		t.Error(err)
	}

	blockBookWrap := blockbook.NewBlockBookWrapper(coinConfig.Info.Blockbook)

	//** get the balance, check if its > 0 or less than the amount
	info, err := blockBookWrap.GetEthAddress(account.Address.Hex())
	if err != nil {
		t.Error(err)
	}

	nonce, err := strconv.ParseUint(info.Nonce, 0, 64)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(account.Address.Hex(), nonce)
	//** Retrieve information for outputs: out address
	toAddress := common.HexToAddress("0x673153460D01A22F9dAc129F2Ea59be3681921A4")

	//**calculate fee/gas cost, add the amount
	gasLimit := uint64(200000)
	gasStation := GasStation{}
	_ = getJson("https://ethgasstation.info/json/ethgasAPI.json", &gasStation)
	gasPrice := big.NewInt(int64(1000000000 * (gasStation.Average / 10)))

	tokenAddress := common.HexToAddress("0xfab46e002bbf0b4509813474841e0716e6730136")

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID), gasLimit, gasPrice, tokenAddress, toAddress) // 0xa9059cbb

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAddress))

	amount := new(big.Int)
	amount.SetUint64(uint64(10 * (math.Pow(10, 18))))
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount), amount)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	value := big.NewInt(0) // in wei (0 eth)
	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)
	// **sign and send
	signedTx, err := wallet.SignTx(account, tx, nil)
	if err != nil {
		t.Error(err)
	}
	ts := types.Transactions{signedTx}
	rawTxBytes := ts.GetRlp(0)
	rawTxHex := hex.EncodeToString(rawTxBytes)
	fmt.Println(rawTxHex)
	response, err := blockBookWrap.SendTx("0x" + rawTxHex)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(response)

}
