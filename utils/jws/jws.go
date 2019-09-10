package jws

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"gopkg.in/square/go-jose.v2"
)

func DecodeJWS(token string, encodedPubKey string) ([]byte,  error) {
	pubKeyBytes, err := base64.StdEncoding.DecodeString(encodedPubKey)
	if err != nil {
		return nil, err
	}
	rsaPubKey, err := x509.ParsePKCS1PublicKey(pubKeyBytes)
	if err != nil {
		return nil, err
	}
	object, err := jose.ParseSigned(token)
	if err != nil {
		return nil, err
	}
	data, err := object.Verify(&rsaPubKey)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func EncodeJWS(payload interface{}, privkey string) (string, error) {
	privKeyBytes, err := base64.StdEncoding.DecodeString(privkey)
	if err != nil {
		return "", err
	}
	rsaPrivKey, err := x509.ParsePKCS1PrivateKey(privKeyBytes)
	if err != nil {
		return "", err
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.PS512, Key: rsaPrivKey}, nil)
	if err != nil {
		return "", err
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	object, err := signer.Sign(payloadBytes)
	if err != nil {
		return "", err
	}
	serialize, err := object.CompactSerialize()
	if err != nil {
		return "", err
	}
	return serialize, nil
}
