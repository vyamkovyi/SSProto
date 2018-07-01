package main

import (
	"crypto/rand"
	"io/ioutil"
)

func newUUID() ([]byte, error) {
	v := make([]byte, 32)
	_, err := rand.Read(v)
	return v, err
}

func UUID() ([]byte, error) {
	uuidLocation := "config/uuid.bin"
	if fileExists(uuidLocation) {
		return ioutil.ReadFile(uuidLocation)
	} else {
		b, err := newUUID()
		if err != nil {
			return nil, err
		}
		ioutil.WriteFile(uuidLocation, b, 0600)
		return b, nil
	}
}
