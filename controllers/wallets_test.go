package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/plutus/config"
	coinfactory "github.com/grupokindynos/plutus/models/coin-factory"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

var wCtrl = WalletController{}

func init() {
	_ = godotenv.Load()
}

func TestWalletController_GetInfo(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "No-Coin"}}
	wCtrl.GetInfo(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoCoin.Error(), response["error"])
	fmt.Println(response)
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_GetInfo2(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "polis"}}
	wCtrl.GetInfo(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoRpcUserProvided.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_GetWalletInfo(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "No-Coin"}}
	wCtrl.GetWalletInfo(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoCoin.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_GetWalletInfo2(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "polis"}}
	wCtrl.GetWalletInfo(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoRpcUserProvided.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_GetAddress(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "No-Coin"}}
	wCtrl.GetAddress(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoCoin.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_GetAddress2(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "polis"}}
	wCtrl.GetAddress(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoRpcUserProvided.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_GetNodeStatus(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "No-Coin"}}
	wCtrl.GetNodeStatus(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoCoin.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_GetNodeStatus2(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "polis"}}
	wCtrl.GetNodeStatus(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoRpcUserProvided.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_SendToAddress(t *testing.T) {
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
}

func TestWalletController_GetTx(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "No-Coin"}}
	wCtrl.GetTx(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoCoin.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_GetTx2(t *testing.T) {
	// Error cases
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	c.Params = gin.Params{gin.Param{Key: "coin", Value: "polis"}}
	wCtrl.GetTx(c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, config.ErrorNoRpcUserProvided.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestWalletController_RPCClient(t *testing.T) {
	polis, err := coinfactory.GetCoin("polis")
	assert.Nil(t, err)
	rpcClient := wCtrl.RPCClient(polis)
	assert.NotNil(t, rpcClient)
}
