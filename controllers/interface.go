package controllers

type Service interface {
	GetBalance(params Params) (interface{}, error)
	GetAddress(params Params) (interface{}, error)
	SendToAddress(params Params) (interface{}, error)
	ValidateAddress(params Params) (interface{}, error)
	DecodeRawTx(params Params) (interface{}, error)
}
