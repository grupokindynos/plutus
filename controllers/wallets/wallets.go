package wallets

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/plutus/config"
	coinfactory "github.com/grupokindynos/plutus/models/coin-factory"
	"github.com/grupokindynos/plutus/models/responses"
	"github.com/grupokindynos/plutus/models/rpc"
	"github.com/ybbus/jsonrpc"
	"time"
)

type WalletController struct{}

func (wc *WalletController) GetWalletInfo(c *gin.Context) {
	coin := c.Param("coin")
	coinConfig := coinfactory.GetCoin(coin)

	hostStr := coinConfig.User + "@" + coinConfig.Host + ":" + coinConfig.Port
	tunnel := config.NewSSHTunnel(hostStr, config.PrivateKeyFile(coinConfig.PrivKey), "localhost:"+coinConfig.RpcPort)
	go func() {
		_ = tunnel.Start()
	}()
	time.Sleep(100 * time.Millisecond)
	rpcClient := jsonrpc.NewClientWithOpts("http://"+tunnel.Local.String(), &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(coinConfig.RpcUser+":"+coinConfig.RpcPass)),
		},
	})
	res, err := rpcClient.Call("getwalletinfo")
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

func (wc *WalletController) GetInfo(c *gin.Context) {
	coin := c.Param("coin")
	coinConfig := coinfactory.GetCoin(coin)
	hostStr := coinConfig.User + "@" + coinConfig.Host + ":" + coinConfig.Port
	tunnel := config.NewSSHTunnel(hostStr, config.PrivateKeyFile(coinConfig.PrivKey), "localhost:"+coinConfig.RpcPort)
	go func() {
		_ = tunnel.Start()
	}()
	time.Sleep(100 * time.Millisecond)
	rpcClient := jsonrpc.NewClientWithOpts("http://"+tunnel.Local.String(), &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(coinConfig.RpcUser+":"+coinConfig.RpcPass)),
		},
	})
	chainRes, err := rpcClient.Call("getblockchaininfo")
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
	netRes, err := rpcClient.Call("getnetworkinfo")
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

func (wc *WalletController) GetAddress(c *gin.Context) {
	coin := c.Param("coin")
	coinConfig := coinfactory.GetCoin(coin)
	hostStr := coinConfig.User + "@" + coinConfig.Host + ":" + coinConfig.Port
	tunnel := config.NewSSHTunnel(hostStr, config.PrivateKeyFile(coinConfig.PrivKey), "localhost:"+coinConfig.RpcPort)
	go func() {
		_ = tunnel.Start()
	}()
	time.Sleep(100 * time.Millisecond)
	rpcClient := jsonrpc.NewClientWithOpts("http://"+tunnel.Local.String(), &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(coinConfig.RpcUser+":"+coinConfig.RpcPass)),
		},
	})
	res, err := rpcClient.Call("getnewaddress")
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
