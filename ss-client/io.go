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
	"fmt"
	"io"
	"strings"
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

// SendHashListEntry writes serializes hashlist entry to out io.Writer.
func SendHashListEntry(out io.Writer, path string, hash []byte) error {
	err := binary.Write(out, binary.LittleEndian, hash)
	if err != nil {
		return err
	}
	bytesPath := []byte(path)
	err = binary.Write(out, binary.LittleEndian, uint64(len(bytesPath)))
	if err != nil {
		return err
	}
	return binary.Write(out, binary.LittleEndian, bytesPath)
}

// Packet is an update unit that contains file that needs to be updated and some metadata
type Packet struct {
	FilePath string
	Blob     io.Reader
	Size     uint64
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
	err = binary.Read(in, binary.LittleEndian, &res.Size)
	if err != nil {
		return nil, err
	}
	res.Blob = io.LimitReader(in, int64(res.Size))
	return res, nil
}

// WriteTo implements io.WriterTo for Packet. Each packet must be
// written before reading next one.
func (p Packet) WriteTo(w io.Writer) (int64, error) {
	return io.Copy(w, p.Blob)
}

func copyWithProgress(filename string, size uint64, src io.Reader, dst io.Writer) error {
	written := uint64(0)
	buf := make([]byte, 65536) // There is nothing wrong with using big buffers.

	eof := false
	var msgLength int
	for !eof {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += uint64(nw)
			}
			if ew != nil {
				return ew
			}
			if nr != nw {
				return io.ErrShortWrite
			}
		}
		if er != nil {
			if er == io.EOF {
				eof = true
			} else {
				return er
			}
		}

		newstr := fmt.Sprintf("\rReceiving %s (%s of %s, %v%%)...",
			filename, humanReadableSize(written), humanReadableSize(size),
			int(float64(written)/float64(size)*100))
		if len(newstr) < msgLength {
			newstr += strings.Repeat(" ", msgLength-len(newstr))
		}
		msgLength = len(newstr)
		fmt.Print(newstr)
	}
	newstr := fmt.Sprintf("\rReceived %s", filename)
	if msgLength > len(newstr) {
		newstr += strings.Repeat(" ", msgLength-len(newstr))
	}
	msgLength = len(newstr)
	fmt.Print(newstr + "\n")

	return nil
}
