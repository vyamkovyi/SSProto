package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"io/ioutil"
	"log"
	"time"

	"github.com/twstrike/ed448"
	"crypto/tls"
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
		binary.Write(conn, binary.LittleEndian, pv < SSProtoVersion)
	}

	// Expecting 32-bytes long identifier
	data := make([]byte, 32)
	err := binary.Read(conn, binary.LittleEndian, data)
	if err != nil {
		log.Println("Stream error:", err)
		return
	}

	// Sign identifier
	signature, err := SignData(data)
	if err != nil {
		log.Println("Unable to sign received client identifier:", err)
		return
	}

	// Send signature back
	err = binary.Write(conn, binary.LittleEndian, signature)
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

		// Read file
		s, err := ioutil.ReadFile(entry.ServPath)
		if err != nil {
			log.Panicln("Failed to read file", entry.ServPath)
		}

		// Generate and send hash
		err = binary.Write(conn, binary.LittleEndian, entry.Hash)
		if err != nil {
			log.Println("Stream error:", err)
			return
		}

		// Sign hash and send signature
		curve = ed448.NewDecafCurve()
		signature, ok := curve.Sign(privateKey, entry.Hash[:])
		if !ok {
			log.Panicln("Failed to sign hash!")
		}
		err = binary.Write(conn, binary.LittleEndian, signature)
		if err != nil {
			log.Println("Stream error:", err)
			return
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

	log.Println("HWInfo:", baseEncodedID+":"+string(machineData))
	log.Println("Success!")
}
