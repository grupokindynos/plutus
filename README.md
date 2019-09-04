# Plutus
> Plutus is the Greek god of wealth

[![CircleCI](https://circleci.com/gh/grupokindynos/plutus.svg?style=svg)](https://circleci.com/gh/grupokindynos/plutus)
[![codecov](https://codecov.io/gh/grupokindynos/plutus/branch/master/graph/badge.svg)](https://codecov.io/gh/grupokindynos/plutus)[![Go Report](https://goreportcard.com/badge/github.com/grupokindynos/plutus)](https://goreportcard.com/report/github.com/grupokindynos/plutus) 
[![GoDocs](https://godoc.org/github.com/grupokindynos/plutus?status.svg)](http://godoc.org/github.com/grupokindynos/plutus)

Plutus is a microservice API for ultra safe access to multiple cryptocurrency hot-wallets

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/grupokindynos/plutus)

## Building

To run Plutus from the source code, first you need to install golang, follow this guide:
```
https://golang.org/doc/install
```

To run Plutus simply clone de repository:

```
git clone https://github.com/grupokindynos/plutus 
```

Install dependencies
```
go mod download
```

Build it or Run it:
```
go build && ./plutus
```
```
go run main.go
```

Make sure the port is configured under en enviroment variable `PORT=8080`


## API Reference
@TODO

## Testing

Simply run:
```
go test ./...
```

## Contributing

Pull requests accepted.

To add a new coin, you need to add parameters on `models/coin-factory/coins.go` and add the variable to the `Coins` map.

Also, you need to add the environment variables to access the hot-wallet over a ssh tunnel.
Currently every coin uses 9 variables following this structure:

```
{Coin_ticker_uppercase}_IP=
{Coin_ticker_uppercase}_RPC_USER=
{Coin_ticker_uppercase}_RPC_PASS=
{Coin_ticker_uppercase}_RPC_PORT=
{Coin_ticker_uppercase}_SSH_USER=
{Coin_ticker_uppercase}_SSH_PORT=
{Coin_ticker_uppercase}_SSH_PRIVKEY=
{Coin_ticker_uppercase}_EXCHANGE_ADDRESS=
{Coin_ticker_uppercase}_COLD_ADDRESS=
```

The variables can be set using a `.env` or defining the variables on specifically like Docker or Heroku.

The entire description is available on the heroku template `app.json`

Make sure the variables are compatible with current implementation.