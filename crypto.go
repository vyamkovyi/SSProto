package main

import (
	"encoding/base64"

	"github.com/twstrike/ed448"
)

var publicKey [56]byte
var curve ed448.DecafCurve

func LoadKeys() error {
	pubEnc := "X/uLlbPShKzadjbEGjok9fyqeNuVDeG8l6IDcBmxO2MSC2Q82og5cFaY2tGJSaAUmn8nYGmXBEc="
	publicKeySlice, pubErr := base64.StdEncoding.DecodeString(pubEnc)
	if pubErr != nil {
		return pubErr
	}
	copy(publicKey[:], publicKeySlice)
	curve = ed448.NewDecafCurve()
	return nil
}

func Verify(data []byte, signature [112]byte) bool {
	verify, err := curve.Verify(signature, data, publicKey)
	return verify && err == nil
}
