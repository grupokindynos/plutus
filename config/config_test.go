package config

import (
	"crypto/aes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestNewEndpoint(t *testing.T) {
	port := 8080
	endpointStr := "user:password@localhost:" + strconv.Itoa(port)
	endpoint := NewEndpoint(endpointStr)
	assert.Equal(t, "user:password", endpoint.User)
	assert.Equal(t, "localhost", endpoint.Host)
	assert.Equal(t, port, endpoint.Port)
	assert.Equal(t, "localhost:8080", endpoint.String())
}

func TestEncryptionError(t *testing.T) {
	messageStr := "test message encryption"
	// Key size 20
	key := "12345678901112131415"
	_, err := Encrypt([]byte(key), []byte(messageStr))
	assert.NotNil(t, err)
	assert.Equal(t, aes.KeySizeError(len(key)).Error(), err.Error())
}

func TestDecryptionError(t *testing.T) {
	messageEncrypted := "pGoO0Df5u2weI47b4bUUt0cWtULg46ctTbmMLibJ8SVxl16zA1xF"
	key := "12345678901112131415"
	_, err := Decrypt([]byte(key), messageEncrypted)
	assert.NotNil(t, err)
	assert.Equal(t, aes.KeySizeError(len(key)).Error(), err.Error())
}

func TestEncryption(t *testing.T) {
	messageStr := "test message encryption"
	key := "1234567890111213"
	encryptedMsg, err := Encrypt([]byte(key), []byte(messageStr))
	assert.Nil(t, err)
	decryptedMsg, err := Decrypt([]byte(key), encryptedMsg)
	assert.Nil(t, err)
	assert.Equal(t, decryptedMsg, messageStr)
}

func TestGlobalResponseError(t *testing.T) {
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	newErr := errors.New("test error")
	_ = GlobalResponse(nil, newErr, c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Nil(t, response["data"])
	assert.Equal(t, newErr.Error(), response["error"])
	assert.Equal(t, float64(-1), response["status"])
}

func TestGlobalResponseSuccess(t *testing.T) {
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	mockData := "success"
	_ = GlobalResponse(mockData, nil, c)
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Nil(t, response["error"])
	assert.Equal(t, mockData, response["data"])
	assert.Equal(t, float64(1), response["status"])
}
