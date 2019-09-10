package common

type BodyReq struct {
	Payload string `bson:"payload" json:"payload"`
}
