package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/plutus/config"
	"github.com/grupokindynos/plutus/models/blockbook"
	coinfactory "github.com/grupokindynos/plutus/models/coin-factory"
	"github.com/grupokindynos/plutus/models/responses"
	"github.com/grupokindynos/plutus/models/rpc"
	"github.com/ybbus/jsonrpc"
	"io/ioutil"
	"time"
)

type WalletController struct{}

func (w *WalletController) GetWalletInfo(c *gin.Context) {
	coin := c.Param("coin")
	coinConfig, err := coinfactory.GetCoin(coin)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	err = w.CheckConfigs(coinConfig)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	rpcClient := w.RPCClient(coinConfig)
	res, err := rpcClient.Call(coinConfig.RpcMethods.GetWalletInfo)
	if err != nil {
		config.GlobalResponse(nil, config.ErrorRpcConnection, c)
		return
	}
	var WalletInfo rpc.GetWalletInfo
	err = res.GetObject(&WalletInfo)
	if err != nil {
		config.GlobalResponse(nil, config.ErrorRpcDeserialize, c)
		return
	}
	response := responses.Balance{
		Confirmed:   WalletInfo.Balance,
		Unconfirmed: WalletInfo.UnconfirmedBalance,
	}
	config.GlobalResponse(response, err, c)
	return
}

func (w *WalletController) GetInfo(c *gin.Context) {
	coin := c.Param("coin")
	coinConfig, err := coinfactory.GetCoin(coin)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	err = w.CheckConfigs(coinConfig)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	rpcClient := w.RPCClient(coinConfig)
	chainRes, err := rpcClient.Call(coinConfig.RpcMethods.GetBlockchainInfo)
	if err != nil {
		config.GlobalResponse(nil, config.ErrorRpcConnection, c)
		return
	}
	var ChainInfo rpc.GetBlockchainInfo
	err = chainRes.GetObject(&ChainInfo)
	if err != nil {
		config.GlobalResponse(nil, config.ErrorRpcDeserialize, c)
		return
	}
	netRes, err := rpcClient.Call(coinConfig.RpcMethods.GetNetworkInfo)
	if err != nil {
		config.GlobalResponse(nil, config.ErrorRpcConnection, c)
		return
	}
	var NetInfo rpc.GetNetworkInfo
	err = netRes.GetObject(&NetInfo)
	if err != nil {
		config.GlobalResponse(nil, config.ErrorRpcDeserialize, c)
		return
	}
	response := responses.Info{
		Blocks:      ChainInfo.Blocks,
		Headers:     ChainInfo.Headers,
		Chain:       ChainInfo.Chain,
		Protocol:    NetInfo.Protocolversion,
		Version:     NetInfo.Version,
		Subversion:  NetInfo.Subversion,
		Connections: NetInfo.Connections,
	}
	config.GlobalResponse(response, err, c)
	return
}

func (w *WalletController) GetAddress(c *gin.Context) {
	coin := c.Param("coin")
	coinConfig, err := coinfactory.GetCoin(coin)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	err = w.CheckConfigs(coinConfig)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	rpcClient := w.RPCClient(coinConfig)
	res, err := rpcClient.Call(coinConfig.RpcMethods.GetNewAddress, jsonrpc.Params(""))
	if err != nil {
		config.GlobalResponse(nil, config.ErrorRpcConnection, c)
		return
	}
	address, err := res.GetString()
	if err != nil {
		config.GlobalResponse(nil, config.ErrorRpcDeserialize, c)
		return
	}
	config.GlobalResponse(address, err, c)
	return
}

func (w *WalletController) GetNodeStatus(c *gin.Context) {
	coin := c.Param("coin")
	coinConfig, err := coinfactory.GetCoin(coin)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	err = w.CheckConfigs(coinConfig)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	rpcClient := w.RPCClient(coinConfig)
	chainRes, err := rpcClient.Call(coinConfig.RpcMethods.GetBlockchainInfo)
	if err != nil {
		config.GlobalResponse(nil, config.ErrorRpcConnection, c)
		return
	}
	var nodeStatus rpc.GetBlockchainInfo
	err = chainRes.GetObject(&nodeStatus)
	if err != nil {
		config.GlobalResponse(nil, config.ErrorRpcDeserialize, c)
		return
	}
	externalRes, err := config.HttpClient.Get("https://" + coinConfig.ExternalSource + "/api")
	if err != nil {
		config.GlobalResponse(nil, config.ErrorExternalStatusError, c)
		return
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
	response := responses.Status{
		NodeBlocks:      nodeStatus.Blocks,
		NodeHeaders:     nodeStatus.Headers,
		ExternalBlocks:  externalStatus.Backend.Blocks,
		ExternalHeaders: externalStatus.Backend.Headers,
		Synced:          isSynced,
	}
	config.GlobalResponse(response, nil, c)
	return
}

func (w *WalletController) SendToAddress(c *gin.Context) {
	coin := c.Param("coin")
	coinConfig, err := coinfactory.GetCoin(coin)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	err = w.CheckConfigs(coinConfig)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	address := c.Param("address")
	if address == "" {
		config.GlobalResponse(nil, config.ErrorNoAddressSpecified, c)
		return
	}
	amount, ok := c.GetQuery("amount")
	if !ok {
		config.GlobalResponse(nil, config.ErrorNoAmountSpecified, c)
		return
	}
	txid, err := w.Send(coinConfig, address, amount)
	if err != nil {
		config.GlobalResponse(nil, config.ErrorUnableToSend, c)
		return
	}
	config.GlobalResponse(txid, nil, c)
	return
}

func (w *WalletController) SendToColdStorage(c *gin.Context) {
	coin := c.Param("coin")
	coinConfig, err := coinfactory.GetCoin(coin)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	err = w.CheckConfigs(coinConfig)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	amount, ok := c.GetQuery("amount")
	if !ok {
		config.GlobalResponse(nil, config.ErrorNoAmountSpecified, c)
		return
	}
	txid, err := w.Send(coinConfig, coinConfig.ColdAddress, amount)
	if err != nil {
		config.GlobalResponse(nil, config.ErrorUnableToSend, c)
		return
	}
	config.GlobalResponse(txid, nil, c)
	return
}

func (w *WalletController) SendToExchange(c *gin.Context) {
	coin := c.Param("coin")
	coinConfig, err := coinfactory.GetCoin(coin)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	err = w.CheckConfigs(coinConfig)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	amount, ok := c.GetQuery("amount")
	if !ok {
		config.GlobalResponse(nil, config.ErrorNoAmountSpecified, c)
		return
	}
	txid, err := w.Send(coinConfig, coinConfig.ExchangeAddress, amount)
	if err != nil {
		config.GlobalResponse(nil, config.ErrorUnableToSend, c)
		return
	}
	config.GlobalResponse(txid, nil, c)
	return
}

func (w *WalletController) ValidateAddress(c *gin.Context) {
	coin := c.Param("coin")
	address := c.Param("address")
	coinConfig, err := coinfactory.GetCoin(coin)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	err = w.CheckConfigs(coinConfig)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	rpcClient := w.RPCClient(coinConfig)
	res, err := rpcClient.Call(coinConfig.RpcMethods.ValidateAddress, jsonrpc.Params(address))
	if err != nil {
		config.GlobalResponse(nil, config.ErrorUnableToValidateAddress, c)
		return
	}
	var AddressValidation rpc.ValidateAddress
	err = res.GetObject(&AddressValidation)
	if err != nil {
		config.GlobalResponse(nil, config.ErrorRpcDeserialize, c)
		return
	}
	response := responses.Address{
		Valid: AddressValidation.Ismine,
	}
	config.GlobalResponse(response, nil, c)
	return
}

func (w *WalletController) GetTx(c *gin.Context) {
	coin := c.Param("coin")
	txid := c.Param("txid")
	coinConfig, err := coinfactory.GetCoin(coin)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	err = w.CheckConfigs(coinConfig)
	if err != nil {
		config.GlobalResponse(nil, err, c)
		return
	}
	rpcClient := w.RPCClient(coinConfig)
	res, err := rpcClient.Call(coinConfig.RpcMethods.GetRawTransaction, jsonrpc.Params(txid, coinConfig.RpcMethods.GetRawTransactionVerbosity))
	if err != nil {
		config.GlobalResponse(nil, config.ErrorUnableToValidateAddress, c)
		return
	}
	config.GlobalResponse(res.Result, nil, c)
	return
}

func (w *WalletController) RPCClient(coinConfig *coinfactory.Coin) jsonrpc.RPCClient {
	hostStr := coinConfig.User + "@" + coinConfig.Host + ":" + coinConfig.Port
	tunnel := config.NewSSHTunnel(hostStr, config.PrivateKey(coinConfig.PrivKey), "localhost:"+coinConfig.RpcPort)
	go func() {
		// TODO handle tunnel error
		err := tunnel.Start()
		if err != nil {
			fmt.Println(err)
		}
	}()
	time.Sleep(100 * time.Millisecond)
	rpcClient := jsonrpc.NewClientWithOpts("http://"+tunnel.Local.String(), &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(coinConfig.RpcUser+":"+coinConfig.RpcPass)),
		},
	})
	return rpcClient
}

func (w *WalletController) Send(coinConfig *coinfactory.Coin, address string, amount string) (string, error) {
	rpcClient := w.RPCClient(coinConfig)
	chainRes, err := rpcClient.Call(coinConfig.RpcMethods.SendToAddress, jsonrpc.Params(address, amount))
	if err != nil {
		return "", err
	}
	return chainRes.GetString()
}

func (w *WalletController) CheckConfigs(coin *coinfactory.Coin) error {
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
	/*	if coin.ExchangeAddress == "" {
			return config.ErrorNoExchangeAddress
		}
		if coin.ColdAddress == "" {
			return config.ErrorNoColdAddress
		}*/

	return nil
}
