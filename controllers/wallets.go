package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	res "github.com/grupokindynos/common/responses"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
	"github.com/grupokindynos/plutus/config"
	"github.com/grupokindynos/plutus/models/blockbook"
	"github.com/grupokindynos/plutus/models/common"
	"github.com/grupokindynos/plutus/models/responses"
	"github.com/grupokindynos/plutus/models/rpc"
	"github.com/grupokindynos/plutus/utils"
	"github.com/ybbus/jsonrpc"
	"io/ioutil"
	"os"
	"time"
)

type RPCClient jsonrpc.RPCClient

type WalletController struct{}

func (w *WalletController) GetInfo(c *gin.Context) {
	_, err := utils.VerifyHeaderSignature(c)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	coin := c.Param("coin")
	coinConfig, err := coinfactory.GetCoin(coin)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	rpcClient := w.RPCClient(coinConfig)
	chainRes, err := rpcClient.Call(coinConfig.RpcMethods.GetBlockchainInfo)
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorRpcConnection, c)
		return
	}
	var ChainInfo rpc.GetBlockchainInfo
	err = chainRes.GetObject(&ChainInfo)
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorRpcDeserialize, c)
		return
	}
	netRes, err := rpcClient.Call(coinConfig.RpcMethods.GetNetworkInfo)
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorRpcConnection, c)
		return
	}
	var NetInfo rpc.GetNetworkInfo
	err = netRes.GetObject(&NetInfo)
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorRpcDeserialize, c)
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
	header, body, err := mrt.CreateMRTToken("plutus", os.Getenv("MASTER_PASSWORD"), response, os.Getenv("PLUTUS_PRIVATE_KEY"))
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	res.GlobalResponseMRT(header, body, c)
	return
}

func (w *WalletController) GetWalletInfo(c *gin.Context) {
	_, err := utils.VerifyHeaderSignature(c)
	if err != nil {
		res.GlobalResponseNoAuth(c)
		return
	}
	coin := c.Param("coin")
	coinConfig, err := coinfactory.GetCoin(coin)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	rpcClient := w.RPCClient(coinConfig)
	callRes, err := rpcClient.Call(coinConfig.RpcMethods.GetWalletInfo)
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorRpcConnection, c)
		return
	}
	var WalletInfo rpc.GetWalletInfo
	err = callRes.GetObject(&WalletInfo)
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorRpcDeserialize, c)
		return
	}
	response := responses.Balance{
		Confirmed:   WalletInfo.Balance,
		Unconfirmed: WalletInfo.UnconfirmedBalance,
	}
	header, body, err := mrt.CreateMRTToken("plutus", os.Getenv("MASTER_PASSWORD"), response, os.Getenv("PLUTUS_PRIVATE_KEY"))
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	res.GlobalResponseMRT(header, body, c)
	return
}

func (w *WalletController) GetAddress(c *gin.Context) {
	_, err := utils.VerifyHeaderSignature(c)
	if err != nil {
		res.GlobalResponseNoAuth(c)
		return
	}
	coin := c.Param("coin")
	coinConfig, err := coinfactory.GetCoin(coin)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	rpcClient := w.RPCClient(coinConfig)
	callRes, err := rpcClient.Call(coinConfig.RpcMethods.GetNewAddress, jsonrpc.Params(""))
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorRpcConnection, c)
		return
	}
	address, err := callRes.GetString()
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorRpcDeserialize, c)
		return
	}
	header, body, err := mrt.CreateMRTToken("plutus", os.Getenv("MASTER_PASSWORD"), address, os.Getenv("PLUTUS_PRIVATE_KEY"))
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	res.GlobalResponseMRT(header, body, c)
	return
}

func (w *WalletController) GetNodeStatus(c *gin.Context) {
	_, err := utils.VerifyHeaderSignature(c)
	if err != nil {
		res.GlobalResponseNoAuth(c)
		return
	}
	coin := c.Param("coin")
	coinConfig, err := coinfactory.GetCoin(coin)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	rpcClient := w.RPCClient(coinConfig)
	chainRes, err := rpcClient.Call(coinConfig.RpcMethods.GetBlockchainInfo)
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorRpcConnection, c)
		return
	}
	var nodeStatus rpc.GetBlockchainInfo
	err = chainRes.GetObject(&nodeStatus)
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorRpcDeserialize, c)
		return
	}
	externalRes, err := config.HttpClient.Get("https://" + coinConfig.BlockchainInfo.ExternalSource + "/api")
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorExternalStatusError, c)
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
	header, body, err := mrt.CreateMRTToken("plutus", os.Getenv("MASTER_PASSWORD"), response, os.Getenv("PLUTUS_PRIVATE_KEY"))
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	res.GlobalResponseMRT(header, body, c)
	return
}

func (w *WalletController) SendToAddress(c *gin.Context) {
	servicePubKey, err := utils.VerifyHeaderSignature(c)
	if err != nil {
		res.GlobalResponseNoAuth(c)
		return
	}
	var BodyReq common.BodyReq
	err = c.BindJSON(&BodyReq)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	valid, payload := mvt.VerifyMVTToken(c.GetHeader("service"), BodyReq.Payload, servicePubKey, os.Getenv("MASTER_PASSWORD"))
	if !valid {
		res.GlobalResponseNoAuth(c)
		return
	}
	var SendToAddressData common.SendAddressBodyReq
	err = json.Unmarshal(payload, &SendToAddressData)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	coinConfig, err := coinfactory.GetCoin(SendToAddressData.Coin)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	txid, err := w.Send(coinConfig, SendToAddressData.Address, fmt.Sprintf("%f", SendToAddressData.Amount))
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorUnableToSend, c)
		return
	}
	response := common.ResponseTxid{Txid: txid}
	header, body, err := mrt.CreateMRTToken("plutus", os.Getenv("MASTER_PASSWORD"), response, os.Getenv("PLUTUS_PRIVATE_KEY"))
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	res.GlobalResponseMRT(header, body, c)
	return
}

func (w *WalletController) SendToColdStorage(c *gin.Context) {
	servicePubKey, err := utils.VerifyHeaderSignature(c)
	if err != nil {
		res.GlobalResponseNoAuth(c)
		return
	}
	var BodyReq common.BodyReq
	err = c.BindJSON(&BodyReq)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	valid, payload := mvt.VerifyMVTToken(c.GetHeader("service"), BodyReq.Payload, servicePubKey, os.Getenv("MASTER_PASSWORD"))
	if !valid {
		res.GlobalResponseNoAuth(c)
		return
	}
	var SendToAddressData common.SendAddressInternalBodyReq
	err = json.Unmarshal(payload, &SendToAddressData)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	coinConfig, err := coinfactory.GetCoin(SendToAddressData.Coin)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	txid, err := w.Send(coinConfig, coinConfig.ColdAddress, fmt.Sprintf("%f", SendToAddressData.Amount))
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorUnableToSend, c)
		return
	}
	response := common.ResponseTxid{Txid: txid}
	header, body, err := mrt.CreateMRTToken("plutus", os.Getenv("MASTER_PASSWORD"), response, os.Getenv("PLUTUS_PRIVATE_KEY"))
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	res.GlobalResponseMRT(header, body, c)
	return
}

func (w *WalletController) SendToExchange(c *gin.Context) {
	servicePubKey, err := utils.VerifyHeaderSignature(c)
	if err != nil {
		res.GlobalResponseNoAuth(c)
		return
	}
	var BodyReq common.BodyReq
	err = c.BindJSON(&BodyReq)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	valid, payload := mvt.VerifyMVTToken(c.GetHeader("service"), BodyReq.Payload, servicePubKey, os.Getenv("MASTER_PASSWORD"))
	if !valid {
		res.GlobalResponseNoAuth(c)
		return
	}
	var SendToAddressData common.SendAddressBodyReq
	err = json.Unmarshal(payload, &SendToAddressData)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	coinConfig, err := coinfactory.GetCoin(SendToAddressData.Coin)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	txid, err := w.Send(coinConfig, SendToAddressData.Address, fmt.Sprintf("%f", SendToAddressData.Amount))
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorUnableToSend, c)
		return
	}
	response := common.ResponseTxid{Txid: txid}
	header, body, err := mrt.CreateMRTToken("plutus", os.Getenv("MASTER_PASSWORD"), response, os.Getenv("PLUTUS_PRIVATE_KEY"))
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	res.GlobalResponseMRT(header, body, c)
	return
}

func (w *WalletController) ValidateAddress(c *gin.Context) {
	servicePubKey, err := utils.VerifyHeaderSignature(c)
	if err != nil {
		res.GlobalResponseNoAuth(c)
		return
	}
	var BodyReq common.BodyReq
	err = c.BindJSON(&BodyReq)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	valid, payload := mvt.VerifyMVTToken(c.GetHeader("service"), BodyReq.Payload, servicePubKey, os.Getenv("MASTER_PASSWORD"))
	if !valid {
		res.GlobalResponseNoAuth(c)
		return
	}
	var ValidateAddressData common.AddressValidationBodyReq
	err = json.Unmarshal(payload, &ValidateAddressData)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	coinConfig, err := coinfactory.GetCoin(ValidateAddressData.Coin)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	rpcClient := w.RPCClient(coinConfig)
	resCall, err := rpcClient.Call(coinConfig.RpcMethods.ValidateAddress, jsonrpc.Params(ValidateAddressData.Address))
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorUnableToValidateAddress, c)
		return
	}
	var AddressValidation rpc.ValidateAddress
	err = resCall.GetObject(&AddressValidation)
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorRpcDeserialize, c)
		return
	}
	response := responses.Address{
		Valid: AddressValidation.Ismine,
	}
	header, body, err := mrt.CreateMRTToken("plutus", os.Getenv("MASTER_PASSWORD"), response, os.Getenv("PLUTUS_PRIVATE_KEY"))
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	res.GlobalResponseMRT(header, body, c)
	return
}

func (w *WalletController) GetTx(c *gin.Context) {
	_, err := utils.VerifyHeaderSignature(c)
	if err != nil {
		res.GlobalResponseNoAuth(c)
		return
	}
	coin := c.Param("coin")
	txid := c.Param("txid")
	coinConfig, err := coinfactory.GetCoin(coin)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	err = coinfactory.CheckCoinKeys(coinConfig)
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	rpcClient := w.RPCClient(coinConfig)
	resCall, err := rpcClient.Call(coinConfig.RpcMethods.GetRawTransaction, jsonrpc.Params(txid, coinConfig.RpcMethods.GetRawTransactionVerbosity))
	if err != nil {
		res.GlobalResponseError(nil, config.ErrorUnableToValidateAddress, c)
		return
	}
	header, body, err := mrt.CreateMRTToken("plutus", os.Getenv("MASTER_PASSWORD"), resCall.Result, os.Getenv("PLUTUS_PRIVATE_KEY"))
	if err != nil {
		res.GlobalResponseError(nil, err, c)
		return
	}
	res.GlobalResponseMRT(header, body, c)
	return
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
