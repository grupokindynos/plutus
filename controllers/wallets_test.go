package controllers

import (
	"github.com/grupokindynos/common/coin-factory"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"testing"
)

var wCtrl = WalletController{}

func init() {
	_ = godotenv.Load()
}

var TestParamsError = Params{
	Coin: "No-COIN",
}

var TestParams = Params{
	Coin: "polis",
}

func TestWalletController_GetInfo(t *testing.T) {
	// Error cases
	res, err := wCtrl.GetInfo(TestParamsError)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Equal(t, "coin not available", err.Error())
}

func TestWalletController_GetInfo2(t *testing.T) {
	// Error cases
	res, err := wCtrl.GetInfo(TestParams)
	assert.NotNil(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "missing rpc username", err.Error())
}

func TestWalletController_GetWalletInfo(t *testing.T) {
	// Error cases
	res, err := wCtrl.GetWalletInfo(TestParamsError)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Equal(t, "coin not available", err.Error())
}

func TestWalletController_GetWalletInfo2(t *testing.T) {
	// Error cases
	res, err := wCtrl.GetWalletInfo(TestParams)
	assert.NotNil(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "missing rpc username", err.Error())
}

func TestWalletController_GetAddress(t *testing.T) {
	// Error cases
	res, err := wCtrl.GetAddress(TestParamsError)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Equal(t, "coin not available", err.Error())
}

func TestWalletController_GetAddress2(t *testing.T) {
	// Error cases
	res, err := wCtrl.GetAddress(TestParams)
	assert.NotNil(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "missing rpc username", err.Error())
}

func TestWalletController_GetNodeStatus(t *testing.T) {
	// Error cases
	res, err := wCtrl.GetNodeStatus(TestParamsError)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Equal(t, "coin not available", err.Error())
}

func TestWalletController_GetNodeStatus2(t *testing.T) {
	// Error cases
	res, err := wCtrl.GetNodeStatus(TestParams)
	assert.NotNil(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "missing rpc username", err.Error())
}

// TODO migrate tests
/*func TestWalletController_SendToAddress(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "No-Coin"}}
	wCtrl.SendToAddress(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoCoin.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_SendToAddress2(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "polis"}}
	wCtrl.SendToAddress(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoRpcUserProvided.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_SendToColdStorage(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "No-Coin"}}
	wCtrl.SendToColdStorage(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoCoin.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_SendToColdStorage2(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "polis"}}
	wCtrl.SendToColdStorage(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoRpcUserProvided.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_SendToExchange(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "No-Coin"}}
	wCtrl.SendToExchange(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoCoin.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_SendToExchange2(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "polis"}}
	wCtrl.SendToExchange(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoRpcUserProvided.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_ValidateAddress(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "No-Coin"}}
	wCtrl.ValidateAddress(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoCoin.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_ValidateAddress2(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "polis"}}
	wCtrl.ValidateAddress(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoRpcUserProvided.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}*/

func TestWalletController_GetTx(t *testing.T) {
	// Error cases
	res, err := wCtrl.GetTx(TestParamsError)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Equal(t, "coin not available", err.Error())
}

func TestWalletController_GetTx2(t *testing.T) {
	// Error cases
	res, err := wCtrl.GetTx(TestParams)
	assert.NotNil(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "missing rpc username", err.Error())
}

func TestWalletController_RPCClient(t *testing.T) {
	polis, err := coinfactory.GetCoin("polis")
	assert.Nil(t, err)
	tunnel := getNewTunnel(polis)
	rpcClient := wCtrl.RPCClient(polis, tunnel)
	assert.NotNil(t, rpcClient)
}
