package main

import (
	"encoding/base64"

	"github.com/twstrike/ed448"
)

var publicKey [56]byte
var curve ed448.DecafCurve

func LoadKeys() error {
	pubEnc := "Krjuy0kHp0r1ADKDWAjo+odqLNlVCnb4MAmvbyWMezEt3C18LCWVPVV2d8ggYc0f83p7Wyqd3TU="
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
