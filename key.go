package main

import (
	"encoding/base64"
	"log"
)

func loadKeys() [56]byte {
	pubEnc := "P4urJQKQBRqXIreG/ZBK606F14YeewR0pHcjHfdnTMDp58cLmE76rEhv3MF1+NeWYhxvOqfvvxU="
	publicKeySlice, pubErr := base64.StdEncoding.DecodeString(pubEnc)
	if pubErr != nil {
		log.Panicln("Invalid key file content!")
	}
	publicKey := [56]byte{}
	copy(publicKey[:], publicKeySlice)
	return publicKey
}
