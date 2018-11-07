// io.go - communication with the update server
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
	"encoding/binary"
	"encoding/json"
	"io"
)

// WriteHWInfo writes machine information in form of JSON to given Writer.
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

// SendHashListEntry serializes packet to a given pipe
func SendHashListEntry(pipe io.ReadWriter, path string, hash []byte) (bool, error) {
	err := binary.Write(pipe, binary.LittleEndian, hash)
	if err != nil {
		return false, err
	}
	bytesPath := []byte(path)
	err = binary.Write(pipe, binary.LittleEndian, uint64(len(bytesPath)))
	if err != nil {
		return false, err
	}
	err = binary.Write(pipe, binary.LittleEndian, bytesPath)
	if err != nil {
		return false, err
	}
	resp := false
	err = binary.Read(pipe, binary.LittleEndian, &resp)
	return resp, err
}

// Packet is an update unit that contains file that needs to be updated and some metadata
type Packet struct {
	FilePath string
	Blob     io.Reader
}

// ReadPacket deserializes packet structure from a binary stream
func ReadPacket(in io.Reader) (*Packet, error) {
	res := new(Packet)
	var size uint64
	err := binary.Read(in, binary.LittleEndian, &size)
	if err != nil {
		return nil, err
	}
	pathBytes := make([]byte, size)
	err = binary.Read(in, binary.LittleEndian, pathBytes)
	if err != nil {
		return nil, err
	}
	res.FilePath = string(pathBytes)
	size = uint64(0)
	err = binary.Read(in, binary.LittleEndian, &size)
	if err != nil {
		return nil, err
	}
	res.Blob = io.LimitReader(in, int64(size))
	return res, nil
}

// WriteTo implements io.WriterTo for Packet. Each packet must be
// written before reading next one.
func (p Packet) WriteTo(w io.Writer) (int64, error) {
	return io.Copy(w, p.Blob)
}
