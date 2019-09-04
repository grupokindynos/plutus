package coinfactory

import (
	"github.com/grupokindynos/plutus/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetCoin(t *testing.T) {
	coin, err := GetCoin("BTC")
	assert.Nil(t, err)
	assert.IsType(t, &Coin{}, coin)
}

func TestGetCoinError(t *testing.T) {
	coin, err := GetCoin("NO-COIN")
	assert.NotNil(t, err)
	assert.Nil(t, coin)
	assert.Equal(t, config.ErrorNoCoin, err)
}

func TestCheckCoinConfigs(t *testing.T) {
	coin := &Coin{
		Tag: "BTC",
	}
	err := CheckCoinConfigs(coin)
	assert.NotNil(t, err)
	assert.Equal(t, config.ErrorNoRpcUserProvided, err)
	coin.RpcUser = "MockUser"
	err = CheckCoinConfigs(coin)
	assert.NotNil(t, err)
	assert.Equal(t, config.ErrorNoRpcPassProvided, err)
	coin.RpcPass = "MockPass"
	err = CheckCoinConfigs(coin)
	assert.NotNil(t, err)
	assert.Equal(t, config.ErrorNoRpcPortProvided, err)
	coin.RpcPort = "33344"
	err = CheckCoinConfigs(coin)
	assert.NotNil(t, err)
	assert.Equal(t, config.ErrorNoHostIPProvided, err)
	coin.Host = "1.1.1.1"
	err = CheckCoinConfigs(coin)
	assert.NotNil(t, err)
	assert.Equal(t, config.ErrorNoHostPortProvided, err)
	coin.Port = "12312"
	err = CheckCoinConfigs(coin)
	assert.NotNil(t, err)
	assert.Equal(t, config.ErrorNoHostUserProvided, err)
	coin.User = "cronos"
	err = CheckCoinConfigs(coin)
	assert.NotNil(t, err)
	assert.Equal(t, config.ErrorNoAuthMethodProvided, err)
	coin.PrivKey = "mockPrivKey"
	err = CheckCoinConfigs(coin)
	assert.NotNil(t, err)
	assert.Equal(t, config.ErrorNoExchangeAddress, err)
	coin.ExchangeAddress = "RandomAddr"
	err = CheckCoinConfigs(coin)
	assert.NotNil(t, err)
	assert.Equal(t, config.ErrorNoColdAddress, err)
	coin.ColdAddress = "RandomAddr"
	err = CheckCoinConfigs(coin)
	assert.Nil(t, err)
}
