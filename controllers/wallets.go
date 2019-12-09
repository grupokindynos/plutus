package controllers

import (
	"encoding/json"
	"fmt"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/plutus"
	"github.com/grupokindynos/plutus/models"
	"github.com/ybbus/jsonrpc"
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

func (w *WalletController) GetWalletInfo(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	fmt.Print(coinConfig)
	return nil, nil
}

func (w *WalletController) GetAddress(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	fmt.Print(coinConfig)
	return nil, nil
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
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		return nil, err
	}
	txid, err := w.Send(coinConfig, SendToAddressData.Address, SendToAddressData.Amount)
	if err != nil {
		return nil, err
	}
	return txid, nil
}

func (w *WalletController) SendToColdStorage(params Params) (interface{}, error) {
	var SendToAddressData plutus.SendAddressInternalBodyReq
	err := json.Unmarshal(params.Body, &SendToAddressData)
	if err != nil {
		return nil, err
	}
	coinConfig, err := coinfactory.GetCoin(SendToAddressData.Coin)
	if err != nil {
		return nil, err
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		return nil, err
	}
	txid, err := w.Send(coinConfig, coinConfig.ColdAddress, SendToAddressData.Amount)
	if err != nil {
		return nil, err
	}
	response := models.ResponseTxid{Txid: txid}
	return response, nil
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
	err = coinfactory.CheckCoinKeys(coinConfig)
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
