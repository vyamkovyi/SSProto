package main

import (
	"net"
	"time"
	"encoding/binary"
	"log"
	"github.com/twstrike/ed448"
	"bytes"
	"io/ioutil"
	"encoding/base64"
	"strings"
)

func (s *Service) serve(conn *net.TCPConn) {
	defer conn.Close()
	defer s.wg.Done()
	conn.SetDeadline(time.Now().Add(time.Second * 600))
	var size uint64

	// Protocol version
	{
		var pv uint8
		err := binary.Read(conn, binary.LittleEndian, &pv)
		if err != nil {
			log.Println("Stream error:", err.Error())
			return
		}
		binary.Write(conn, binary.LittleEndian, pv < SSProtoVersion)
	}

	// Expecting 32-bytes long identifier
	data := make([]byte, 32)
	err := binary.Read(conn, binary.LittleEndian, data)
	if err != nil {
		log.Println("Stream error:", err.Error())
		return
	}

	// Sign identifier
	curve := ed448.NewDecafCurve()
	signature, ok := curve.Sign(privateKey, data)
	if !ok {
		log.Println("Unable to sign received client identifier!")
		return
	}

	// Send signature back
	err = binary.Write(conn, binary.LittleEndian, signature)
	if err != nil {
		log.Println("Stream error:", err.Error())
		return
	}

	// Record machine data if it wasn't recorded yet
	baseEncodedID := base64.StdEncoding.EncodeToString(data)
	var machineData []byte

	if machineExists(baseEncodedID) {
		log.Println("Rejecting connection - already served today.")
		err = binary.Write(conn, binary.LittleEndian, false)
		if err != nil {
			log.Println("Stream error:", err.Error())
		}
		return
	}

	err = binary.Write(conn, binary.LittleEndian, true)
	if err != nil {
		log.Println("Stream error:", err.Error())
		return
	}
	binary.Read(conn, binary.LittleEndian, &size)
	machineData = make([]byte, size)
	err = binary.Read(conn, binary.LittleEndian, machineData)
	if err != nil {
		log.Println("Stream error:", err.Error())
		return
	}

	tempMap := make(map[[32]byte]string)

	// Get hashes from client and create an intersection
	for {
		// Expect file hash
		var hash [32]byte
		err = binary.Read(conn, binary.LittleEndian, &hash)
		if err != nil {
			log.Println("Stream error:", err.Error())
			return
		}

		if bytes.Equal(hash[:], make([]byte, 32)) {
			break
		}

		contains := false
		if v, ok := filesMap[hash]; ok {
			contains = true
			tempMap[hash] = v
		}

		err := binary.Write(conn, binary.LittleEndian, contains)
		if err != nil {
			log.Println("Stream error:", err.Error())
			return
		}
	}

	// Remove difference from server files to create a list of mods that we need to send
	tempMap2 := make(map[[32]byte]string)
	for k, v := range filesMap {
		if _, ok := tempMap[k]; ok {
			continue
		}
		tempMap2[k] = v
	}

	for k, v := range tempMap2 {
		// Read file
		s, err := ioutil.ReadFile(v)
		if err != nil {
			log.Panicln("Failed to read file", v)
		}

		// Generate and send hash
		err = binary.Write(conn, binary.LittleEndian, k)
		if err != nil {
			log.Println("Stream error:", err.Error())
			return
		}

		// Sign hash and send signature
		curve = ed448.NewDecafCurve()
		signature, ok := curve.Sign(privateKey, k[:])
		if !ok {
			log.Panicln("Failed to sign hash!")
		}
		err = binary.Write(conn, binary.LittleEndian, signature)
		if err != nil {
			log.Println("Stream error:", err.Error())
			return
		}

		// File path
		var pathBytes []byte
		pathBytes = []byte(strings.TrimPrefix(v, "client/"))
		err = binary.Write(conn, binary.LittleEndian, pathBytes)
		if err != nil {
			log.Println("Stream error:", err.Error())
			return
		}
		binary.Write(conn, binary.LittleEndian, byte(0))

		// Size of file
		size = uint64(len(s))
		err = binary.Write(conn, binary.LittleEndian, size)
		if err != nil {
			log.Println("Stream error:", err.Error())
			return
		}

		// File blob
		err = binary.Write(conn, binary.LittleEndian, s)
		if err != nil {
			log.Println("Stream error:", err.Error())
			return
		}
	}

	log.Println("HWInfo:", baseEncodedID + ":" + string(machineData))
	log.Println("Success!")
}