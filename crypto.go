package main

import (
	"encoding/base64"

	"crypto/tls"
	"crypto/x509"

	"github.com/twstrike/ed448"
)

var publicKey [56]byte
var curve ed448.DecafCurve
var cert = `-----BEGIN CERTIFICATE-----
MIICtzCCAg2gAwIBAgIJAIaJfVK6hNQFMAoGCCqGSM49BAMCMGcxCzAJBgNVBAYT
AkRFMREwDwYDVQQKDAhIZXhhbWluZTERMA8GA1UECwwISGV4YW1pbmUxETAPBgNV
BAMMCEhleGF3b2xmMR8wHQYJKoZIhvcNAQkBFhBoZXhhd29sZkBjb2NrLmxpMB4X
DTE4MDYzMDIwMTQzOVoXDTIwMDYyOTIwMTQzOVowZzELMAkGA1UEBhMCREUxETAP
BgNVBAoMCEhleGFtaW5lMREwDwYDVQQLDAhIZXhhbWluZTERMA8GA1UEAwwISGV4
YXdvbGYxHzAdBgkqhkiG9w0BCQEWEGhleGF3b2xmQGNvY2subGkwgacwEAYHKoZI
zj0CAQYFK4EEACcDgZIABAMxXo9xiXQ2ljv+zmjuJGZBMnWF1BXSPUzDQs/5hsO+
WxUMWH8Wr/Y2Y/tydG8jX6RbXDD2Fyn0cyLjJxiZ+tcaIYmV45wK8ASirB59h238
lCeczEGC5ax8kaxkYktxMO5MRmilMOQwGKdC5qfJj/qAiu9y5+OV3lijnqgl87aX
gF5pAlGDhw05q3V96qzY4KNTMFEwHQYDVR0OBBYEFFPydedpJgQW8laSmYi6sbXV
Zo+dMB8GA1UdIwQYMBaAFFPydedpJgQW8laSmYi6sbXVZo+dMA8GA1UdEwEB/wQF
MAMBAf8wCgYIKoZIzj0EAwIDgZcAMIGTAkcbc5cAtlqPXE8zqd05QA9b5C++iUUb
iIlVa8isvwiBpIS1PX0Eqq0T1OI0rhZVfzpL9uStEpcpFLwkdUcXDT8CjlNT++Fp
XAJIA+r7LpcUV2qgmSGENzQ2Xv9Gr15/fFAJFHblKS1DvBHGPNvUCoWHDKp0yryP
+vOGTCiuGDgP42P/iv9yV+vAg9qAQi1IuKY8
-----END CERTIFICATE-----
`
var conf tls.Config

func LoadKeys() error {
	pubEnc := "X/uLlbPShKzadjbEGjok9fyqeNuVDeG8l6IDcBmxO2MSC2Q82og5cFaY2tGJSaAUmn8nYGmXBEc="
	publicKeySlice, pubErr := base64.StdEncoding.DecodeString(pubEnc)
	if pubErr != nil {
		return pubErr
	}
	copy(publicKey[:], publicKeySlice)
	curve = ed448.NewDecafCurve()
	certs := x509.NewCertPool()
	certs.AppendCertsFromPEM([]byte(cert))
	conf = tls.Config{
		RootCAs:    certs,
		ServerName: TargetHost,
	}
	return nil
}

func Verify(data []byte, signature [112]byte) bool {
	verify, err := curve.Verify(signature, data, publicKey)
	return verify && err == nil
}
