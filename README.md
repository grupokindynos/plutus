# Plutus
> Plutus is the Greek god of wealth

[![CircleCI](https://circleci.com/gh/grupokindynos/plutus.svg?style=svg)](https://circleci.com/gh/grupokindynos/plutus)
[![codecov](https://codecov.io/gh/grupokindynos/plutus/branch/master/graph/badge.svg)](https://codecov.io/gh/grupokindynos/plutus)[![Go Report](https://goreportcard.com/badge/github.com/grupokindynos/plutus)](https://goreportcard.com/report/github.com/grupokindynos/plutus) 
[![GoDocs](https://godoc.org/github.com/grupokindynos/plutus?status.svg)](http://godoc.org/github.com/grupokindynos/plutus)

Plutus is a microservice API for ultra safe access to multiple cryptocurrency hot-wallets


## Deploy

#### Heroku

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/grupokindynos/plutus)

#### Docker

To deploy to docker, simply pull the image
```
docker pull kindynos/plutus:latest
```
Create a new `.env` file with all the necesarry enviroment variables defined on `app.json`

Run the docker image
```
docker run -p 8080:8080 --env-file .env kindynos/plutus:latest 
```

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

> All the routes are password protected with AUTH_USERNAME and AUTH_PASSWORD set on environment variables

### Get hot-wallet status:

Retrieves the hot-wallet status

**GET method:**

```
https://plutus-wallets.herokuapp.com/status/:coin
```

This will retrieve the current blocks and headers compared with an external source from the hot-wallet
```
{
  "data": {
    "node_blocks": 428371,
    "node_headers": 428371,
    "external_blocks": 428371,
    "external_headers": 428371,
    "synced": true
  },
  "status": 1
}
```

### Get hot-wallet information:

Retrieves the hot-wallet information

**GET method:**

```
https://plutus-wallets.herokuapp.com/info/:coin
```

This will retrieve the current information from the hot-wallet
```
{
  "data": {
    "blocks": 428371,
    "headers": 428371,
    "chain": "main",
    "protocol": 70217,
    "version": 1041800,
    "subversion": "/Polis Core:1.4.18/",
    "connections": 8
  },
  "status": 1
}
```

### Get hot-wallet balance:

Retrieves the hot-wallet balance

**GET method:**

```
https://plutus-wallets.herokuapp.com/balance/:coin
```

This will retrieve the current hot-wallet balance

```
{
  "data": {
    "confirmed": 0.99998065,
    "unconfirmed": 0
  },
  "status": 1
}
```

### Get hot-wallet txid:

Retrieves the hot-wallet transaction from a specified coin

**GET method:**

```
https://plutus-wallets.herokuapp.com/tx/:coin/:txid
```

This will retrieve the current transaction information directly from the node

```
{
  "data": {
    "blockhash": "c1833fbfadce7e992dd693ec1be57095dba13332c99060b7ff1e9e2f428bc621",
    "blocktime": 1567619196,
    "confirmations": 1,
    "height": 428385,
    "hex": "0200000001bb6206d650c3e6a6131254e0464e2ddf9eaa5b51fe21553bc6e28c65a5cd36560100000000ffffffff030000000000000000009b1c48981b0000001976a91414e9dbf14a7ca713084b62fc46b6973ec5b3946088ac0020aa44000000001976a914e1794c6c34efe3a15467898ea8d440fc76b5de2a88ac00000000",
    "locktime": 0,
    "size": 128,
    "time": 1567619196,
    "txid": "b3a76649c60103e98e817da006fcbf429b251cd107f81683f4b628f5a299a2df",
    "version": 2,
    "vin": [
      {
        "scriptSig": {
          "asm": "",
          "hex": ""
        },
        "sequence": 4294967295,
        "txid": "5636cda5658ce2c63b5521fe515baa9edf2d4e46e0541213a6e6c350d60662bb",
        "vout": 1
      }
    ],
    "vout": [
      {
        "n": 0,
        "scriptPubKey": {
          "asm": "",
          "hex": "",
          "type": "nonstandard"
        },
        "value": 0,
        "valueSat": 0
      },
      {
        "n": 1,
        "scriptPubKey": {
          "addresses": [
            "PAVkpdQ7CsgJ7w9mxDdGMStqq3Z7P6BGnp"
          ],
          "asm": "OP_DUP OP_HASH160 14e9dbf14a7ca713084b62fc46b6973ec5b39460 OP_EQUALVERIFY OP_CHECKSIG",
          "hex": "76a91414e9dbf14a7ca713084b62fc46b6973ec5b3946088ac",
          "reqSigs": 1,
          "type": "pubkeyhash"
        },
        "value": 1185.18979739,
        "valueSat": 118518979739
      },
      {
        "n": 2,
        "scriptPubKey": {
          "addresses": [
            "PV9NW71AsR922TnCNKTQ1sg1FLMP76V9HL"
          ],
          "asm": "OP_DUP OP_HASH160 e1794c6c34efe3a15467898ea8d440fc76b5de2a OP_EQUALVERIFY OP_CHECKSIG",
          "hex": "76a914e1794c6c34efe3a15467898ea8d440fc76b5de2a88ac",
          "reqSigs": 1,
          "type": "pubkeyhash"
        },
        "value": 11.52,
        "valueSat": 1152000000
      }
    ]
  },
  "status": 1
}
```

### Get hot-wallet deposit address:

Retrieves a hot-wallet deposit address

**GET method:**

```
https://plutus-wallets.herokuapp.com/address/:coin
```

This will retrieve a new address to deposit to the hot-wallet

```
{
  "data": "PLFJALB3PzmyVmfmttWLhLmWj4FmLR51Xt",
  "status": 1
}
```

### Validate a hot-wallet deposit address:

Retrieves wherever the address is valid as a deposit address

**GET method:**

```
https://plutus-wallets.herokuapp.com/validate/address/:coin
```

This will return address validity

```
{
  "data": {
    "valid": true
  },
  "status": 1
}
```

### Send to an address from the hot-wallet:

Sends the specified amount to another address

**GET method:**

```
https://plutus-wallets.herokuapp.com/send/address/:coin/:address?amount=
```

This will return the txid if it is successful

```
{
  "data": "1e7b7c37b6d13477c717802a09f0a85cb9b3dac7ae7a3cab82d3c7490671510d",
  "status": 1
}
```

### Send to cold storage from the hot-wallet:

Sends the specified amount to the cold storage

**GET method:**

```
https://plutus-wallets.herokuapp.com/send/cold/:coin?amount=
```

This will return the txid if it is successful

```
{
  "data": "45ffe106ef7e27e099d05cce74f366db14e3d8e2049fad39dfdd9a08434eb329",
  "status": 1
}
```

### Send to exchange from the hot-wallet:

Sends the specified amount to the exchange

**GET method:**

```
https://plutus-wallets.herokuapp.com/send/exchange/:coin?amount=
```

This will return the txid if it is successful

```
{
  "data": "cbeee709647a4a69a8b310ef2634c5ec35f142b545f30db7a0c09fe5c7bcec84",
  "status": 1
}
```

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