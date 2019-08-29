package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

func main() {
	file := flag.String("file", "", "private key location string")
	password := flag.String("key", "", "encryption key (max length 32 bytes)")
	flag.Parse()
	if *file == "" {
		log.Fatal(errors.New("please set a file path"))
	}
	if *password == "" {
		log.Fatal(errors.New("please set a password"))
	}
	fileBytes, err := ioutil.ReadFile(*file)
	if err != nil {
		log.Fatal(err)
	}
	encrypted, err := encrypt([]byte(*password), fileBytes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Encrypted Key")
	fmt.Println(encrypted)
}

func encrypt(key []byte, message []byte) (encmess string, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	cipherText := make([]byte, aes.BlockSize+len(message))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], message)
	encmess = base64.StdEncoding.EncodeToString(cipherText)
	return
}