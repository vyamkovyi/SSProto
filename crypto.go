package main

import (
	"log"
	"encoding/base64"
	"os"
	"github.com/twstrike/ed448"
	"bufio"
)

var privateKey [144]byte
var publicKey [56]byte

func makeKeys() {
	f, err := os.Create("sss.key")
	if err != nil {
		log.Panicln("Cannot create key file!", err.Error())
	}
	os.Chmod("sss.key", 0600)
	curve := ed448.NewDecafCurve()
	var ok bool
	privateKey, publicKey, ok = curve.GenerateKeys()
	if !ok {
		log.Panicln("Unable to generate keys!")
	}
	privEnc := base64.StdEncoding.EncodeToString(privateKey[:])
	pubEnc := base64.StdEncoding.EncodeToString(publicKey[:])
	f.WriteString(privEnc + "\n")
	f.WriteString(pubEnc + "\n")
	f.Sync()
	f.Close()
}

func loadKeys() {
	f, err := os.Open("sss.key")
	if err != nil {
		log.Panicln("Cannot read keys!", err.Error())
	}
	reader := bufio.NewReader(f)
	privEnc, privErr := reader.ReadString('\n')
	pubEnc, pubErr := reader.ReadString('\n')
	if pubErr != nil || privErr != nil {
		log.Panicln("Invalid key file content!")
	}
	f.Close()
	var publicKeySlice, privateKeySlice []byte
	privateKeySlice, privErr = base64.StdEncoding.DecodeString(privEnc)
	publicKeySlice, pubErr = base64.StdEncoding.DecodeString(pubEnc)
	if privErr != nil || pubErr != nil {
		log.Panicln("Invalid key file content!")
	}
	copy(privateKey[:], privateKeySlice)
	copy(publicKey[:], publicKeySlice)
}
