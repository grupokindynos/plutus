package common

type BodyReq struct {
	Payload string `bson:"payload" json:"payload"`
}

type AddressValidationBodyReq struct {
	Address string `json:"address"`
	Coin    string `json:"coin"`
}

type SendAddressBodyReq struct {
	Address string  `json:"address"`
	Coin    string  `json:"coin"`
	Amount  float64 `json:"amount"`
}

type SendAddressInternalBodyReq struct {
	Coin   string  `json:"coin"`
	Amount float64 `json:"amount"`
}

type ResponseTxid struct {
	Txid string `json:"txid"`
}