package config

import (
	"crypto/aes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gliderlabs/ssh"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
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


func TestPrivateKey(t *testing.T) {
	// Private Key parsing
	err := os.Setenv("KEY_PASSWORD", "rb8L7BKBDG9shnB6j8EPG67MwHaWC8Rw")
	privateKeyStr := "56Rl0LvcyZxr+av38CQD9P3xUeZ/CVlsxTAlMIlCrQ5oGWd6pG3VaHbfkwqNzG94rQF7p/eY3tvOPLyBZvuScVFgFnWZ5gxqp+aHVY5ltoTcGycolfwziXFlS2TJ3t5v/dgTw7hOhVKKiaDZIRonc5dt6I7exw5S0sg7TIiLgiOOqAkSE5xa0GwY3p6+N3jS5bjQHRWg4KHK65tu34AyQadhTbOm9l4dkgkoeHTE6G28nWv7iwsJRJHS3wWAgp0BftRayKSgqOxjMQOHHc8ithzWKLEsrwx/F1aYfX/F2kR6g0NK8Uf91dGJ5LHfOK1TmD/J"
	assert.Nil(t, err)
	auth := PrivateKey(privateKeyStr)
	assert.NotNil(t, auth)
}

func TestPrivateKeyErr(t *testing.T) {
	// Private Key parsing
	privateKeyStr := "56Rl0LvcyZxr+av38CQD9P3xUeZ/CVlsxTAlMIlCrQ5oGWd6pG3VaHbfkwqNzG94rQF7p/eY3tvOPLyBZvuScVFgFnWZ5gxqp+aHVY5ltoTcGycolfwziXFlS2TJ3t5v/dgTw7hOhVKKiaDZIRonc5dt6I7exw5S0sg7TIiLgiOOqAkSE5xa0GwY3p6+N3jS5bjQHRWg4KHK65tu34AyQadhTbOm9l4dkgkoeHTE6G28nWv7iwsJRJHS3wWAgp0BftRayKSgqOxjMQOHHc8ithzWKLEsrwx/F1aYfX/F2kR6g0NK8Uf91dGJ5LHfOK1TmD/J"
	auth := PrivateKey(privateKeyStr)
	assert.Nil(t, auth)
}

func TestNewSSHTunnel(t *testing.T) {
	go func() {
		_ = ssh.ListenAndServe(":2222", nil)
	}()
	tunnel := NewSSHTunnel("localhost:2222", nil,"localhost")
	assert.NotNil(t, tunnel)
	go func() {
		err := tunnel.Start()
		assert.Nil(t, err)
	}()
	time.Sleep(100 * time.Millisecond)

}