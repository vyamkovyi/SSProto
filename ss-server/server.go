// server.go - does all the SSProto magic ✨
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
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"io/ioutil"
	"log"
	"time"
)

func (s *Service) serve(conn *tls.Conn) {
	defer conn.Close()
	defer s.wg.Done()
	conn.SetDeadline(time.Now().Add(time.Second * 300))
	var size uint64

	// Protocol version
	{
		var pv uint8
		err := binary.Read(conn, binary.LittleEndian, &pv)
		if err != nil {
			log.Println("Stream error:", err)
			return
		}
		binary.Write(conn, binary.LittleEndian, SSProtoVersion)
	}

	// Force pending reindexing if any so we will not
	// send newer version of file when we have only
	// hash of older version.
	filesMapLock.Lock()
	if reindexRequired {
		log.Println("Reindexing files...")
		ListFiles()
		log.Println("Reindexing done")
		reindexRequired = false
		reindexTimer.Stop()
	}
	filesMapLock.Unlock()

	// Expecting 32-bytes long identifier
	data := make([]byte, 32)
	err := binary.Read(conn, binary.LittleEndian, data)
	if err != nil {
		log.Println("Stream error:", err)
		return
	}

	// Record machine data if it wasn't recorded yet
	baseEncodedID := base64.StdEncoding.EncodeToString(data)
	var machineData []byte

	if machineExists(baseEncodedID) {
		log.Println("Rejecting connection - already served today.")
		err = binary.Write(conn, binary.LittleEndian, false)
		if err != nil {
			log.Println("Stream error:", err)
		}
		return
	}

	err = binary.Write(conn, binary.LittleEndian, true)
	if err != nil {
		log.Println("Stream error:", err)
		return
	}
	binary.Read(conn, binary.LittleEndian, &size)
	machineData = make([]byte, size)
	err = binary.Read(conn, binary.LittleEndian, machineData)
	if err != nil {
		log.Println("Stream error:", err)
		return
	}

	clientFiles := make(map[[32]byte]string)
	var clientList []string

	filesMapLock.RLock()
	defer filesMapLock.RUnlock()

	// Get hashes from client and create an intersection
	for {
		// Expect file hash
		var hash [32]byte
		err = binary.Read(conn, binary.LittleEndian, &hash)
		if err != nil {
			log.Println("Stream error:", err)
			return
		}

		if bytes.Equal(hash[:], make([]byte, 32)) {
			break
		}

		// Expect size of file path string
		err = binary.Read(conn, binary.LittleEndian, &size)
		if err != nil {
			log.Println("Stream error:", err)
			return
		}

		// Expect file path
		data = make([]byte, size)
		err = binary.Read(conn, binary.LittleEndian, data)
		if err != nil {
			log.Println("Stream error:", err)
			return
		}

		// Construct client files list
		clientList = append(clientList, string(data))

		// Create intersection of client and server maps
		contains := false
		if v, ok := filesMap[hash]; ok {
			contains = true
			clientFiles[hash] = v.ServPath
		}

		// Answer if file is valid
		err := binary.Write(conn, binary.LittleEndian, contains)
		if err != nil {
			log.Println("Stream error:", err)
			return
		}
	}

	// Remove difference from server files to create a list of mods that we need to send
	changes := make(map[[32]byte]IndexedFile)
	for k, v := range filesMap {
		if _, ok := clientFiles[k]; ok {
			continue
		}
		changes[k] = v
	}

	for _, entry := range changes {
		skip := false
		for _, clientFile := range clientList {
			if clientFile == entry.ClientPath && entry.ShouldNotReplace {
				skip = true
			}
		}
		if skip {
			continue
		}

		// Read file to memory
		s, err := ioutil.ReadFile(entry.ServPath)
		if err != nil {
			log.Panicln("Failed to read file", entry.ServPath)
		}

		// Size of file path
		err = binary.Write(conn, binary.LittleEndian, uint64(len([]byte(entry.ClientPath))))
		if err != nil {
			log.Println("Stream error:", err)
		}

		// File path
		err = binary.Write(conn, binary.LittleEndian, []byte(entry.ClientPath))
		if err != nil {
			log.Println("Stream error:", err)
			return
		}

		// Size of file
		size = uint64(len(s))
		err = binary.Write(conn, binary.LittleEndian, size)
		if err != nil {
			log.Println("Stream error:", err)
			return
		}

		// File blob
		err = binary.Write(conn, binary.LittleEndian, s)
		if err != nil {
			log.Println("Stream error:", err)
			return
		}

	}

	// Logging virtual memory statistics received from the client to the log file
	log.Println("HWInfo:", baseEncodedID+":"+string(machineData))
	log.Println("Success!")
}
