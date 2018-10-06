// crypt.go - loads private keys from filesystem and signs data.
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
	"encoding/base64"
	"os"
	"github.com/twstrike/ed448"
	"bufio"
	"errors"
)

var privateKey [144]byte
var publicKey [56]byte
var curve ed448.DecafCurve

func MakeKeys() error {
	f, err := os.Create("ss.key")
	if err != nil {
		return err
	}
	os.Chmod("ss.key", 0600)
	curve = ed448.NewDecafCurve()
	var ok bool
	privateKey, publicKey, ok = curve.GenerateKeys()
	if !ok {
		return errors.New("unable to generate keys")
	}
	privEnc := base64.StdEncoding.EncodeToString(privateKey[:])
	pubEnc := base64.StdEncoding.EncodeToString(publicKey[:])
	f.WriteString(privEnc + "\n")
	f.WriteString(pubEnc + "\n")
	f.Sync()
	f.Close()
	return nil
}

func LoadKeys() error {
	curve = ed448.NewDecafCurve()
	f, err := os.Open("ss.key")
	if err != nil {
		return err
	}
	reader := bufio.NewReader(f)
	privEnc, privErr := reader.ReadString('\n')
	pubEnc, pubErr := reader.ReadString('\n')
	if pubErr != nil || privErr != nil {
		return errors.New("invalid key file content:" + pubErr.Error())
	}
	f.Close()
	var publicKeySlice, privateKeySlice []byte
	privateKeySlice, privErr = base64.StdEncoding.DecodeString(privEnc)
	publicKeySlice, pubErr = base64.StdEncoding.DecodeString(pubEnc)
	if privErr != nil || pubErr != nil {
		return errors.New("invalid key file content:" + pubErr.Error())
	}
	copy(privateKey[:], privateKeySlice)
	copy(publicKey[:], publicKeySlice)
	return nil
}

func SignData(data []byte) ([112]byte, error) {
	signature, ok := curve.Sign(privateKey, data)
	var err error = nil
	if !ok {
		err = errors.New("unable to sign data")
	}
	return signature, err
}
