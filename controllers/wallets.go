package controllers

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/grupokindynos/plutus/models"
	"io/ioutil"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/plutus"
	"github.com/grupokindynos/plutus/config"
	"github.com/ybbus/jsonrpc"
)

type Params struct {
	Coin string
	Body []byte
	Txid string
}

var ethAccount = "0x4dc011f9792d18cd67f5afa4f1678e9c6c4d8e0e"

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
	var ChainInfo models.GetBlockchainInfo
	err = chainRes.GetObject(&ChainInfo)
	if err != nil {
		return nil, config.ErrorRpcDeserialize
	}
	netRes, err := rpcClient.Call(coinConfig.RpcMethods.GetNetworkInfo)
	if err != nil {
		return nil, config.ErrorRpcConnection
	}
	var NetInfo models.GetNetworkInfo
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
	if coinConfig.Tag != "ETH" {
		rpcClient := w.RPCClient(coinConfig)
		callRes, err := rpcClient.Call(coinConfig.RpcMethods.GetWalletInfo)
		if err != nil {
			return nil, config.ErrorRpcConnection
		}
		var WalletInfo models.GetWalletInfo
		err = callRes.GetObject(&WalletInfo)
		if err != nil {
			return nil, config.ErrorRpcDeserialize
		}
		response := plutus.Balance{
			Confirmed:   WalletInfo.Balance,
			Unconfirmed: WalletInfo.UnconfirmedBalance,
		}
		return response, nil
	} else {
		rpcClient := w.RPCClient(coinConfig)
		balanceRes, err := rpcClient.Call(coinConfig.RpcMethods.GetWalletInfo, jsonrpc.Params(ethAccount, "latest"))
		if err != nil {
			return nil, config.ErrorRpcConnection
		}
		strBalance, err := balanceRes.GetString()
		if err != nil {
			return nil, config.ErrorRpcDeserialize
		}
		confirmed, err := hexutil.DecodeUint64(strBalance)
		if err != nil {
			return nil, config.ErrorRpcDeserialize
		}
		response := plutus.Balance{
			Confirmed:   float64(confirmed),
			Unconfirmed: 0,
		}
		return response, nil
	}

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
	if coinConfig.Tag == "ETH" || coinConfig.Tag == "USDT" || coinConfig.Tag == "USDC" || coinConfig.Tag == "TUSD" {
		return ethAccount, nil
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
	var nodeStatus models.GetBlockchainInfo
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
	var externalStatus models.Status
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
		return nil, config.ErrorUnableToSend
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
		return nil, config.ErrorUnableToSend
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
		return nil, config.ErrorUnableToSend
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
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		return nil, err
	}
	rpcClient := w.RPCClient(coinConfig)
	resCall, err := rpcClient.Call(coinConfig.RpcMethods.ValidateAddress, jsonrpc.Params(ValidateAddressData.Address))
	if err != nil {
		return nil, config.ErrorUnableToValidateAddress
	}
	var AddressValidation models.ValidateAddress
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
	var rawTx string
	err = json.Unmarshal(params.Body, &rawTx)
	if err != nil {
		return nil, err
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		return nil, err
	}
	if coinConfig.Tag == "ETH" {
		var tx interface{}
		rawtx, err := hex.DecodeString(rawTx)
		if err != nil {
			return nil, config.ErrorUnmarshal
		}
		err = rlp.DecodeBytes(rawtx, &tx)
		if err != nil {
			return nil, config.ErrorUnmarshal
		}
		return tx, nil
	} else {
		rpcClient := w.RPCClient(coinConfig)
		resCall, err := rpcClient.Call(coinConfig.RpcMethods.DecodeRawTransaction, jsonrpc.Params(rawTx))
		if err != nil {
			return nil, config.ErrorUnmarshal
		}
		return resCall.Result, nil
	}
}

func (w *WalletController) RPCClient(coinConfig *coins.Coin) RPCClient {
	keys := coinConfig.Keys
	hostStr := keys.User + "@" + keys.Host + ":" + keys.Port
	tunnel := config.NewSSHTunnel(hostStr, config.PrivateKey(keys.PrivKey), "localhost:"+keys.RpcPort)
	go func() {
		_ = tunnel.Start()
	}()
	time.Sleep(1000 * time.Millisecond)
	var rpcClient RPCClient
	if coinConfig.Tag == "ETH" {
		rpcClient = jsonrpc.NewClient("http://" + tunnel.Local.String())
	} else {
		rpcClient = jsonrpc.NewClientWithOpts("http://"+tunnel.Local.String(), &jsonrpc.RPCClientOpts{
			HTTPClient: config.HttpClient,
			CustomHeaders: map[string]string{
				"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(keys.RpcUser+":"+keys.RpcPass)),
			},
		})
	}
	return rpcClient
}

func (w *WalletController) Send(coinConfig *coins.Coin, address string, amount float64) (string, error) {
	rpcClient := w.RPCClient(coinConfig)
	chainRes, err := rpcClient.Call(coinConfig.RpcMethods.SendToAddress, jsonrpc.Params(address, amount))
	if err != nil {
		return "", err
	}
	return chainRes.GetString()
}
