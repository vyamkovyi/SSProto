// crypto.go - hardcoded keys and data signature verification
// Copyright (c) 2018  Hexawolf
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
// of the Software, and to permit persons to whom the Software is furnished to do
// so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
package main

import (
	"crypto/tls"
	"crypto/x509"
	"strings"

	"github.com/denisbrodbeck/machineid"
)

var conf tls.Config

// Both variables are set by build script.
var certEnc, keyEnc string

func LoadKeys() error {
	certs := x509.NewCertPool()
	cert := "-----BEGIN CERTIFICATE-----\n" + certEnc + "\n-----END CERTIFICATE-----"
	certs.AppendCertsFromPEM([]byte(cert))
	conf = tls.Config{
		RootCAs: certs,
		// Extract domain from targetHost
		ServerName: strings.Split(targetHost, ":")[0],
	}
	return nil
}

func UUID(key string) (string, error) {
	id, err := machineid.ProtectedID(key)
	if err != nil {
		return "", err
	}
	return id, err
}
