package controllers

import (
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/martinboehm/btcd/wire"
	"github.com/martinboehm/btcutil"
	"github.com/martinboehm/btcutil/chaincfg"
	"reflect"
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
	path    uint32
	privKey string
	addr    string
}

var testXpup = []testData{
	{path: 10, addr: "1JsMio1jkCBsneBsQxs1cbpj5BwTnZKS8i", privKey: "L24bVGEqUBnQ1pUxoWGtBordNveobDcdxU3zDvkoVWK5RxrYeTde", coin: coinfactory.Coins["BTC"], xpub: "xpub6CJU1RX4dgbFg44YcVY9ighjvsoh184QVXRaxMnxASiCGw6stuAJTTGyWkvyv7d2HKMz2V9hUFBWfYQCjFZUDrxna82vURQTVwkp69poMhx"},
	{path: 10, addr: "XovGmrsTDduAuUdaAMmTCDJ3yx1c8i7bt5", privKey: "XKfbqtXi1mat3YcJ6imRT4x2mWZ6Majo4AzUBgP8iW9VSF1zdyXb", coin: coinfactory.Coins["DASH"], xpub: "xpub6DUFQg5TVUrFp921QUPbEXKSMVZacvYE4BMtR7vNDhtopXNewqs5zFoftATvsgAexA87pgmPCd2hNXfhWcR9UbkuFkShFWgkPvtzgmJragb"},
	{path: 10, addr: "D8tmU27w7vAyw37KHZbZz1Htsefz9RMoFz", privKey: "L1xLSnxKojULWLV45xU4Wtvu7yv6EQ7CWc55z6JtbXN4Sx3Jw4i5", coin: coinfactory.Coins["DGB"], xpub: "xpub6C3WY4bctXDV2xsKCuw2XApxrD64ACn8bPVwDz5UDDgPgzdBunyxNJyEzp7Qt5sf3YEC8UHezceZt2uixmyxsgx5y6CVvxb4eBwJhyCNXZC"},
	//{path: 10, addr: "FpZhQAurnQM5LUZy9KxUmAfzY9P8qqC696", privKey: "L26v51gauTKN6iGDSD6QgCVXxmnA3YvyHByPGp3YejhJq42fJipc",	coin: coinfactory.Coins["GRS"], xpub: "xpub6ChEMNHYLbLEv84XhwyJfjNjNGifp54pion7fDJRf2oRJUMq9NP3mwxZfXDKYZaEWN5MYBE1PFmbNfUBNgJiezXwWjnh6KAW2LYf5U5AbvH"},
	{path: 10, addr: "a2LxezdJMB86HA4WgHJM2WaAcapEPcnSGn", privKey: "Y8izsjrgrwYh6jaXWn9tqUn4wc66GntRTw7mta2z5HNY12QL8xqu", coin: coinfactory.Coins["XZC"], xpub: "xpub6CJYXujizwxUUb6LLfEXyMTjJpqoVoUT24H2dGuGEBtj7tJtXgJLjZA1D3u7TAnYtAQd1AuAfF8UXtQhW6YxwwQ6dDe8MEdTkLPYGSRJ2B2"},
	{path: 10, addr: "LhGukzG7eKsR2MRDUCLWVgqEGsetnM55QK", privKey: "TAxdgg1MXmUdvhoUpX5NA3deShB4z9PVrhbBQxNVDhey5Keh1WHk", coin: coinfactory.Coins["LTC"], xpub: "xpub6CXM64Woerkp9gpYkSekzHkDbyusdCGzeZNgW7qE7Coa3gJBCdwSrhznMmpH2RyVFwyDxTMBQexPFMMg7J3ujZzoMFLtiK6aVMWVxAUtAA7"},
}

func TestXpubGeneration(t *testing.T) {
	for _, test := range testXpup {
		test.coin.Mnemonic = testMnemonic
		acc, err := getAccFromMnemonic(test.coin, true)
		if err != nil {
			panic(err)
		}
		accPub, err := acc.Neuter()
		if err != nil {
			panic(err)
		}
		equalXpub := reflect.DeepEqual(test.xpub, accPub.String())
		if !equalXpub {
			t.Error("addr doesn't match for " + test.coin.Info.Tag + " expected: " + test.xpub + "got: " + accPub.String())
		}
		pubKeyHash, err := getPubKeyHashFromPath(acc, test.coin, test.path)
		if err != nil {
			panic(err)
		}
		equalAddr := reflect.DeepEqual(test.addr, pubKeyHash)
		if !equalAddr {
			t.Error("addr doesn't match for " + test.coin.Info.Tag + " expected: " + test.addr + "got: " + pubKeyHash)
		}
		privKey, err := getPrivKeyFromPath(acc, test.path)
		if err != nil {
			panic(err)
		}
		wif, err := btcutil.NewWIF(privKey, test.coin.NetParams, true)
		if err != nil {
			panic(err)
		}
		equalPrivKey := reflect.DeepEqual(test.privKey, wif.String())
		if !equalPrivKey {
			t.Error("privKey doesn't match for " + test.coin.Info.Tag + " expected: " + test.privKey + "got: " + wif.String())
		}
	}
}
