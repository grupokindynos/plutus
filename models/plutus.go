package models

type BodyReq struct {
	Payload string `bson:"payload" json:"payload"`
}

type AddressValidationBodyReq struct {
	Address string `json:"address"`
	Coin    string `json:"coin"`
}

type TxValidationBodyReq struct {
	Coin    string  `json:"coin"`
	RawTx   string  `json:"raw_tx"`
	Amount  float64 `json:"amount"`
	Address string  `json:"address"`
}

type ResponseTxid struct {
	Txid string `json:"txid"`
}

type AddrInfo struct {
	Addr string
	Path int
}
