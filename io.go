package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"io"

	"os"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func WriteHWInfo(out io.Writer) error {
	b, err := json.Marshal(GetMachineInfo())
	if err != nil {
		return err
	}
	err = binary.Write(out, binary.LittleEndian, uint64(len(b)))
	if err != nil {
		return err
	}
	_, err = out.Write(b)
	return err
}

func readString(in io.Reader) (string, error) {
	var buf []byte
	b := make([]byte, 1)
	for {
		_, err := in.Read(b)
		if err != nil {
			return "", err
		}

		if b[0] == byte(0x00) {
			break
		}

		buf = append(buf, b[0])
	}
	return string(buf), nil
}

func WriteHashList(in map[string][]byte, pipe io.ReadWriter) (map[string]bool,
															  error) {
	res := make(map[string]bool)
	for k, v := range in {
		_, err := pipe.Write(v)
		if err != nil {
			return nil, err
		}
		resp := false
		err = binary.Read(pipe, binary.LittleEndian, &resp)
		if err != nil {
			return nil, err
		}
		res[k] = resp
	}
	zeroes := [32]byte{}
	_, err := pipe.Write(zeroes[:])
	return res, err
}

type Packet struct {
	Hash      [32]byte
	Signature [112]byte
	FilePath  string
	Blob      []byte
}

func ReadPacket(in io.Reader) (*Packet, error) {
	res := new(Packet)
	_, err := in.Read(res.Hash[:])
	if err != nil {
		return nil, err
	}
	_, err = in.Read(res.Signature[:])
	if err != nil {
		return nil, err
	}
	res.FilePath, err = readString(in)
	if err != nil {
		return nil, err
	}
	size := uint64(0)
	err = binary.Read(in, binary.LittleEndian, &size)
	if err != nil {
		return nil, err
	}
	res.Blob = make([]byte, size)
	err = binary.Read(in, binary.LittleEndian, &res.Blob)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (p Packet) Verify() bool {
	sum := sha256.Sum256(p.Blob)
	if sum != p.Hash {
		return false
	}
	// crypto.go
	isValid := Verify(sum[:], p.Signature)
	return isValid
}
