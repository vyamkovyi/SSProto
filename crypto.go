package main

import (
	"encoding/base64"

	"crypto/tls"
	"crypto/x509"

	"github.com/twstrike/ed448"
)

var publicKey [56]byte
var curve ed448.DecafCurve
var conf tls.Config

// Both variables are set by build script.
var certEnc, keyEnc string

func LoadKeys() error {
	publicKeySlice, pubErr := base64.StdEncoding.DecodeString(keyEnc)
	if pubErr != nil {
		return pubErr
	}
	copy(publicKey[:], publicKeySlice)
	curve = ed448.NewDecafCurve()

	certs := x509.NewCertPool()
	cert := "-----BEGIN CERTIFICATE-----\n" + certEnc + "\n-----END CERTIFICATE-----"
	certs.AppendCertsFromPEM([]byte(cert))
	conf = tls.Config{
		RootCAs:    certs,
		ServerName: targetHost,
	}
	return nil
}

func Verify(data []byte, signature [112]byte) bool {
	verify, err := curve.Verify(signature, data, publicKey)
	return verify && err == nil
}
