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

	"crypto/rand"
	"io/ioutil"
	"strings"
)

var conf tls.Config

// Both variables are set by build script.
var certEnc, keyEnc string

// LoadKeys deserializes certificate stored in memory.
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

func newUUID() ([]byte, error) {
	v := make([]byte, 32)
	_, err := rand.Read(v)
	return v, err
}

// UUID tries to load from config/uuid.bin or generate a new random sequence of 32 bytes. This
// sequence is used for client identification.
func UUID() ([]byte, error) {
	uuidLocation := "config/uuid.bin"
	if fileExists(uuidLocation) {
		return ioutil.ReadFile(uuidLocation)
	}
	b, err := newUUID()
	if err != nil {
		return nil, err
	}
	ioutil.WriteFile(uuidLocation, b, 0600)
	return b, nil
}
