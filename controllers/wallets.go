package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/plutus"
	"github.com/grupokindynos/plutus/config"
	"github.com/grupokindynos/plutus/models/blockbook"
	"github.com/grupokindynos/plutus/models/common"
	"github.com/grupokindynos/plutus/models/rpc"
	"github.com/ybbus/jsonrpc"
)

type Params struct {
	Coin string
	Body []byte
	Txid string
}

type RPCClient jsonrpc.RPCClient

type WalletController struct{}

func (w *WalletController) GetInfo(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		return nil, err
	}
	rpcClient := w.RPCClient(coinConfig)
	chainRes, err := rpcClient.Call(coinConfig.RpcMethods.GetBlockchainInfo)
	if err != nil {
		return nil, config.ErrorRpcConnection
	}
	var ChainInfo rpc.GetBlockchainInfo
	err = chainRes.GetObject(&ChainInfo)
	if err != nil {
		return nil, config.ErrorRpcDeserialize
	}
	netRes, err := rpcClient.Call(coinConfig.RpcMethods.GetNetworkInfo)
	if err != nil {
		return nil, config.ErrorRpcConnection
	}
	var NetInfo rpc.GetNetworkInfo
	err = netRes.GetObject(&NetInfo)
	if err != nil {
		return nil, config.ErrorRpcDeserialize
	}
	response := plutus.Info{
		Blocks:      ChainInfo.Blocks,
		Headers:     ChainInfo.Headers,
		Chain:       ChainInfo.Chain,
		Protocol:    NetInfo.Protocolversion,
		Version:     NetInfo.Version,
		SubVersion:  NetInfo.Subversion,
		Connections: NetInfo.Connections,
	}
	return response, nil
}

func (w *WalletController) GetWalletInfo(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		return nil, err
	}
	rpcClient := w.RPCClient(coinConfig)
	callRes, err := rpcClient.Call(coinConfig.RpcMethods.GetWalletInfo)
	if err != nil {
		return nil, config.ErrorRpcConnection
	}
	var WalletInfo rpc.GetWalletInfo
	err = callRes.GetObject(&WalletInfo)
	if err != nil {
		return nil, config.ErrorRpcDeserialize
	}
	response := plutus.Balance{
		Confirmed:   WalletInfo.Balance,
		Unconfirmed: WalletInfo.UnconfirmedBalance,
	}
	return response, nil
}

func (w *WalletController) GetAddress(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		return nil, err
	}
	rpcClient := w.RPCClient(coinConfig)
	callRes, err := rpcClient.Call(coinConfig.RpcMethods.GetNewAddress, jsonrpc.Params(""))
	if err != nil {
		return nil, config.ErrorRpcConnection
	}
	address, err := callRes.GetString()
	if err != nil {
		return nil, config.ErrorRpcDeserialize
	}
	return address, nil
}

func (w *WalletController) GetNodeStatus(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		return nil, err
	}
	rpcClient := w.RPCClient(coinConfig)
	chainRes, err := rpcClient.Call(coinConfig.RpcMethods.GetBlockchainInfo)
	if err != nil {
		return nil, config.ErrorRpcConnection
	}
	var nodeStatus rpc.GetBlockchainInfo
	err = chainRes.GetObject(&nodeStatus)
	if err != nil {
		return nil, config.ErrorRpcDeserialize
	}
	externalRes, err := config.HttpClient.Get("https://" + coinConfig.BlockchainInfo.ExternalSource + "/api")
	if err != nil {
		return nil, config.ErrorRpcConnection
	}
	defer func() {
		_ = externalRes.Body.Close()
	}()
	contents, _ := ioutil.ReadAll(externalRes.Body)
	var externalStatus blockbook.Status
	err = json.Unmarshal(contents, &externalStatus)
	isSynced := false
	if nodeStatus.Blocks == externalStatus.Backend.Blocks && nodeStatus.Headers == externalStatus.Backend.Headers {
		isSynced = true
	}
	response := plutus.Status{
		Blocks:          nodeStatus.Blocks,
		Headers:         nodeStatus.Headers,
		ExternalBlocks:  externalStatus.Backend.Blocks,
		ExternalHeaders: externalStatus.Backend.Headers,
		SyncStatus:      isSynced,
	}
	return response, nil
}

func (w *WalletController) SendToAddress(params Params) (interface{}, error) {
	var SendToAddressData common.SendAddressBodyReq
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
	txid, err := w.Send(coinConfig, SendToAddressData.Address, fmt.Sprintf("%f", SendToAddressData.Amount))
	if err != nil {
		return nil, config.ErrorUnableToSend
	}
	response := common.ResponseTxid{Txid: txid}
	return response, nil
}

func (w *WalletController) SendToColdStorage(params Params) (interface{}, error) {
	var SendToAddressData common.SendAddressInternalBodyReq
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
	txid, err := w.Send(coinConfig, coinConfig.ColdAddress, fmt.Sprintf("%f", SendToAddressData.Amount))
	if err != nil {
		return nil, config.ErrorUnableToSend
	}
	response := common.ResponseTxid{Txid: txid}
	return response, nil
}

func (w *WalletController) SendToExchange(params Params) (interface{}, error) {
	var SendToAddressData common.SendAddressBodyReq
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
	txid, err := w.Send(coinConfig, SendToAddressData.Address, fmt.Sprintf("%f", SendToAddressData.Amount))
	if err != nil {
		return nil, config.ErrorUnableToSend
	}
	response := common.ResponseTxid{Txid: txid}
	return response, nil
}

func (w *WalletController) ValidateAddress(params Params) (interface{}, error) {
	var ValidateAddressData common.AddressValidationBodyReq
	err := json.Unmarshal(params.Body, &ValidateAddressData)
	if err != nil {
		return nil, err
	}
	coinConfig, err := coinfactory.GetCoin(ValidateAddressData.Coin)
	if err != nil {
		return nil, err
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		return nil, err
	}
	rpcClient := w.RPCClient(coinConfig)
	resCall, err := rpcClient.Call(coinConfig.RpcMethods.ValidateAddress, jsonrpc.Params(ValidateAddressData.Address))
	if err != nil {
		return nil, config.ErrorUnableToValidateAddress
	}
	var AddressValidation rpc.ValidateAddress
	err = resCall.GetObject(&AddressValidation)
	if err != nil {
		return nil, config.ErrorRpcDeserialize
	}
	response := plutus.Address{
		Valid: AddressValidation.Ismine,
	}
	return response, nil
}

func (w *WalletController) GetTx(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)
	if err != nil {
		return nil, err
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		return nil, err
	}
	rpcClient := w.RPCClient(coinConfig)
	resCall, err := rpcClient.Call(coinConfig.RpcMethods.GetRawTransaction, jsonrpc.Params(params.Txid, coinConfig.RpcMethods.GetRawTransactionVerbosity))
	if err != nil {
		return nil, config.ErrorUnableToValidateAddress
	}
	return resCall.Result, nil
}

func (w *WalletController) DecodeRawTX(params Params) (interface{}, error) {
	coinConfig, err := coinfactory.GetCoin(params.Coin)

	var txStr string
	err = json.Unmarshal(params.Body, &txStr)
	if err != nil {
		return nil, err
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		return nil, err
	}
	rpcClient := w.RPCClient(coinConfig)
	resCall, err := rpcClient.Call(coinConfig.RpcMethods.DecodeRawTransaction, jsonrpc.Params(txStr))

	if err != nil {
		return nil, config.ErrorUnableToValidateAddress
	}
	return resCall.Result, nil
}

func (w *WalletController) RPCClient(coinConfig *coins.Coin) RPCClient {
	keys := coinConfig.Keys
	hostStr := keys.User + "@" + keys.Host + ":" + keys.Port
	tunnel := config.NewSSHTunnel(hostStr, config.PrivateKey(keys.PrivKey), "localhost:"+keys.RpcPort)
	go func() {
		_ = tunnel.Start()
	}()
	time.Sleep(100 * time.Millisecond)
	rpcClient := jsonrpc.NewClientWithOpts("http://"+tunnel.Local.String(), &jsonrpc.RPCClientOpts{
		HTTPClient: config.HttpClient,
		CustomHeaders: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(keys.RpcUser+":"+keys.RpcPass)),
		},
	})
	return rpcClient
}

func (w *WalletController) Send(coinConfig *coins.Coin, address string, amount string) (string, error) {
	rpcClient := w.RPCClient(coinConfig)
	chainRes, err := rpcClient.Call(coinConfig.RpcMethods.SendToAddress, jsonrpc.Params(address, amount))
	if err != nil {
		return "", err
	}
	return chainRes.GetString()
}
