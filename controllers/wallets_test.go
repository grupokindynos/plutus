package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/plutus/config"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

var wCtrl = WalletController{}

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
}
