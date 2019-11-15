package models

type BodyReq struct {
	Payload string `bson:"payload" json:"payload"`
}

type AddressValidationBodyReq struct {
	Address string `json:"address"`
	Coin    string `json:"coin"`
}

type ResponseTxid struct {
	Txid string `json:"txid"`
}

type TxDeserialize struct {
	Inputs  []Input `json:"input"`
	Outputs []Input `json:"output"`
}

type Input struct {
}

type Output struct {
}
