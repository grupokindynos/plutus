# Plutus
> Plutus is the Greek god of wealth

![Actions](https://github.com/grupokindynos/plutus/workflows/Plutus/badge.svg)
[![codecov](https://codecov.io/gh/grupokindynos/plutus/branch/master/graph/badge.svg)](https://codecov.io/gh/grupokindynos/plutus)
[![Go Report](https://goreportcard.com/badge/github.com/grupokindynos/plutus)](https://goreportcard.com/report/github.com/grupokindynos/plutus) 
[![GoDocs](https://godoc.org/github.com/grupokindynos/plutus?status.svg)](http://godoc.org/github.com/grupokindynos/plutus)

Plutus is a microservice API for for safe access to hot-wallets based on mnemonic phrases.


## Deploy

#### Docker

To deploy to docker, simply pull the image
```
docker pull kindynos/plutus:latest
```
Create a new `.env` file with all the necessary environment variables

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

Documentation: [API Reference](https://documenter.getpostman.com/view/4345063/SVfUs7CX?version=latest)

## Testing

Simply run:
```
go test ./...
```

## Contributing

To contribute to this repository, please fork it, create a new branch and submit a pull request.