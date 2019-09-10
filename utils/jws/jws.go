package jws

func VerifyJWS(token string, signature string) bool {
	return false
}

func DecodeJWS(token string, signature string) interface{} {
	return nil
}

func EncodeJWS(payload interface{}, privkey string) string {
	return ""
}
