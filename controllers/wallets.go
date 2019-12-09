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
	"github.com/ybbus/jsonrpc"
	"os"
	"strconv"
)

type Params struct {
	Coin string
	Body []byte
	Txid string
}

type ERC20Call struct {
	To   string `json:"to"`
	Data string `json:"data"`
}

var ethAccount = "0x4dc011f9792d18cd67f5afa4f1678e9c6c4d8e0e"

type RPCClient jsonrpc.RPCClient

type WalletController struct{}

func (w *WalletController) GetBalance(params Params) (interface{}, error) {
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

func (w *WalletController) GetAddress(params Params) (interface{}, error) {
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
	// Get the last used address
	blockBookWrap, err := blockbook.NewBlockBookWrapper(coinConfig.BlockExplorer)
	if err != nil {
		return nil, err
	}
	info, err := blockBookWrap.GetXpub(acc.String())
	if err != nil {
		return nil, err
	}
	// Create a new xpub and derive the address from the hdwallet
	directExtended, err := acc.Child(0)
	if err != nil {
		return nil, err
	}
	addrExtPub, err := directExtended.Child(uint32(info.UsedTokens + 1))
	if err != nil {
		return nil, err
	}
	addr, err := addrExtPub.Address(coinConfig.NetParams)
	if err != nil {
		return nil, err
	}
	return addr.String(), nil
}

func (w *WalletController) SendToAddress(params Params) (interface{}, error) {
	var SendToAddressData plutus.SendAddressBodyReq
	err := json.Unmarshal(params.Body, &SendToAddressData)
	if err != nil {
		return nil, err
	}
	coinConfig, err := coinfactory.GetCoin(SendToAddressData.Coin)
	if err != nil {
		return nil, err
	}
	txid, err := w.Send(coinConfig, SendToAddressData.Address, SendToAddressData.Amount)
	if err != nil {
		return nil, err
	}
	return txid, nil
}

func (w *WalletController) SendToExchange(params Params) (interface{}, error) {
	var SendToAddressData plutus.SendAddressBodyReq
	err := json.Unmarshal(params.Body, &SendToAddressData)
	if err != nil {
		return nil, err
	}
	coinConfig, err := coinfactory.GetCoin(SendToAddressData.Coin)
	if err != nil {
		return nil, err
	}
	txid, err := w.Send(coinConfig, SendToAddressData.Address, SendToAddressData.Amount)
	if err != nil {
		return nil, err
	}
	response := models.ResponseTxid{Txid: txid}
	return response, nil
}

func (w *WalletController) ValidateAddress(params Params) (interface{}, error) {
	var ValidateAddressData models.AddressValidationBodyReq
	err := json.Unmarshal(params.Body, &ValidateAddressData)
	if err != nil {
		return nil, err
	}
	coinConfig, err := coinfactory.GetCoin(ValidateAddressData.Coin)
	if err != nil {
		return nil, err
	}
	fmt.Print(coinConfig)
	return nil, nil
}

func (w *WalletController) GetTx(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	fmt.Print(coinConfig)
	return nil, nil
}

func (w *WalletController) DecodeRawTX(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	fmt.Print(coinConfig)
	return nil, nil
}

func (w *WalletController) Send(coinConfig *coins.Coin, address string, amount float64) (string, error) {
	fmt.Print(coinConfig)
	return "", nil
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
