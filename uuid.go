package main

import (
	"io/ioutil"
	"math/rand"
	"os"
)

func newUUID() ([]byte, error) {
	v := make([]byte, 32)
	_, err := rand.Read(v)
	return v, err
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
