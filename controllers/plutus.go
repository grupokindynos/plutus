package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/grupokindynos/common/blockbook"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/plutus"
	"github.com/grupokindynos/plutus/models"
	"github.com/tyler-smith/go-bip39"
	"os"
	"strconv"
)

type Params struct {
	Coin string
	Body []byte
	Txid string
}

var ethAccount = "0x4dc011f9792d18cd67f5afa4f1678e9c6c4d8e0e"

const addrGap = 50

type AddrInfo struct {
	LastUsed int
	AddrInfo []models.AddrInfo
}

type Controller struct {
	Address map[string]AddrInfo
}

func (c *Controller) GetBalance(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	acc, err := getAccFromMnemonic(coinConfig)
	if err != nil {
		return nil, err
	}
	blockBookWrap, err := blockbook.NewBlockBookWrapper(coinConfig.BlockExplorer)
	if err != nil {
		return nil, err
	}
	info, err := blockBookWrap.GetXpub(acc.String())
	if err != nil {
		return nil, err
	}
	confirmed, err := strconv.ParseFloat(info.Balance, 64)
	if err != nil {
		return nil, err
	}
	unconfirmed, err := strconv.ParseFloat(info.UnconfirmedBalance, 64)
	if err != nil {
		return nil, err
	}
	response := plutus.Balance{
		Confirmed:   confirmed,
		Unconfirmed: unconfirmed,
	}
	return response, nil
}

func (c *Controller) GetAddress(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	if coinConfig.Token || coinConfig.Tag == "ETH" {
		return ethAccount, nil
	}
	acc, err := getAccFromMnemonic(coinConfig)
	if err != nil {
		return nil, err
	}
	// Create a new xpub and derive the address from the hdwallet
	directExtended, err := acc.Child(0)
	if err != nil {
		return nil, err
	}
	addrExtPub, err := directExtended.Child(uint32(c.Address[coinConfig.Tag].LastUsed + 1))
	if err != nil {
		return nil, err
	}
	addr, err := addrExtPub.Address(coinConfig.NetParams)
	if err != nil {
		return nil, err
	}
	newAddrInfo := AddrInfo{
		LastUsed: c.Address[coinConfig.Tag].LastUsed + 1,
		AddrInfo: c.Address[coinConfig.Tag].AddrInfo,
	}
	newAddrInfo.AddrInfo = append(newAddrInfo.AddrInfo, models.AddrInfo{
		Addr: addr.String(), Path: c.Address[coinConfig.Tag].LastUsed + 1,
	})
	c.Address[coinConfig.Tag] = newAddrInfo
	return addr.String(), nil
}

func (c *Controller) SendToAddress(params Params) (interface{}, error) {
	var SendToAddressData plutus.SendAddressBodyReq
	err := json.Unmarshal(params.Body, &SendToAddressData)
	if err != nil {
		return nil, err
	}
	coinConfig, err := coinfactory.GetCoin(SendToAddressData.Coin)
	if err != nil {
		return nil, err
	}
	fmt.Println(coinConfig)
	return nil, nil
}

func (c *Controller) ValidateAddress(params Params) (interface{}, error) {
	var ValidateAddressData models.AddressValidationBodyReq
	err := json.Unmarshal(params.Body, &ValidateAddressData)
	if err != nil {
		return nil, err
	}
	coinConfig, err := coinfactory.GetCoin(ValidateAddressData.Coin)
	if err != nil {
		return nil, err
	}
	var isMine bool
	for _, addr := range c.Address[coinConfig.Tag].AddrInfo {
		if addr.Addr == ValidateAddressData.Address {
			isMine = true
		}
	}
	return isMine, nil
}

func (c *Controller) DecodeRawTX(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	fmt.Print(coinConfig)
	return nil, nil
}

func (c *Controller) getAddr(coinConfig *coins.Coin) error {
	acc, err := getAccFromMnemonic(coinConfig)
	if err != nil {
		return err
	}
	blockBookWrap, err := blockbook.NewBlockBookWrapper(coinConfig.BlockExplorer)
	if err != nil {
		return err
	}
	info, err := blockBookWrap.GetXpub(acc.String())
	if err != nil {
		return err
	}
	var addrInfoSlice []models.AddrInfo
	for i := info.UsedTokens; i < info.UsedTokens+addrGap; i++ {
		directExtended, err := acc.Child(0)
		if err != nil {
			return err
		}
		addrExtPub, err := directExtended.Child(uint32(i))
		if err != nil {
			return err
		}
		addr, err := addrExtPub.Address(coinConfig.NetParams)
		if err != nil {
			return err
		}
		addrInfo := models.AddrInfo{Addr: addr.String(), Path: i}
		addrInfoSlice = append(addrInfoSlice, addrInfo)
	}
	c.Address[coinConfig.Tag] = AddrInfo{
		LastUsed: info.UsedTokens,
		AddrInfo: addrInfoSlice,
	}
	return nil
}

func getAccFromMnemonic(coinConfig *coins.Coin) (*hdkeychain.ExtendedKey, error) {
	if coinConfig.Mnemonic == "" {
		return nil, errors.New("the coin is not available")
	}
	seed := bip39.NewSeed(coinConfig.Mnemonic, os.Getenv("MNEMONIC_PASSWORD"))
	mKey, err := hdkeychain.NewMaster(seed, coinConfig.NetParams)
	if err != nil {
		return nil, err
	}
	purposeChild, err := mKey.Child(hdkeychain.HardenedKeyStart + 44)
	if err != nil {
		return nil, err
	}
	coinType, err := purposeChild.Child(hdkeychain.HardenedKeyStart + coinConfig.NetParams.HDCoinType)
	if err != nil {
		return nil, err
	}
	accChild, err := coinType.Child(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		return nil, err
	}
	return accChild, nil
}

func NewPlutusController() *Controller {
	ctrl := &Controller{
		Address: make(map[string]AddrInfo),
	}
	// Here we handle only active coins
	for _, coin := range coinfactory.Coins {
		coinConf, err := coinfactory.GetCoin(coin.Tag)
		if err != nil {
			panic(err)
		}
		if coin.Tag == "DASH" || coin.Tag == "BTC" || coin.Tag == "POLIS" {
			err := ctrl.getAddr(coinConf)
			if err != nil {
				panic(err)
			}
		}
	}
	return ctrl
}
