package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"

	"runtime"
	"time"
	"fmt"
)

func collectRecurse(root string) ([]string, error) {
	var res []string
	walkfn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		res = append(res, path)
		return nil
	}
	err := filepath.Walk(root, walkfn)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func collectFlat(root string) ([]string, error) {
	var res []string
	dirl, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}
	for _, info := range dirl {
		if info.IsDir() {
			continue
		}
		res = append(res, filepath.Join(root, info.Name()))
	}
	return res, nil
}

func collectHashList() (map[string][]byte, error) {
	res := make(map[string][]byte)

	err := os.MkdirAll("mods", 0770)
	if err != nil {
		return nil, err
	}
	os.MkdirAll("config", 0700)
	if err != nil {
		return nil, err
	}

	var list []string
	mods, err := collectFlat("mods")
	if err != nil {
		return nil, err
	}
	list = append(list, mods...)
	conf, err := collectRecurse("config")
	if err != nil {
		return nil, err
	}
	list = append(list, conf...)

	for _, path := range list {
		//if _, prs := skippedFiles[filepath.ToSlash(path)]; prs {
		//	log.Println(path, "- IGNORED")
		//	continue
		//}

		blob, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(blob)
		res[path] = sum[:]
	}
	return res, nil
}

const SSProtoVersion uint8 = 1

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	log.Println("SSProto version:", SSProtoVersion)
	log.Println("Copyright (C) Hexawolf, foxcpp 2018")

	c, err := net.Dial("tcp", "doggoat.de:48879")
	if err != nil {
		Crash("Unable to connect to the server:", err.Error())
	}
	defer c.Close()

	// Send protocol version and get answer whether we must ask user for update
	{
		err := binary.Write(c, binary.LittleEndian, SSProtoVersion)
		if err != nil {
			Crash("Unable to send SSProto version:", err.Error())
		}
		var answer bool
		err = binary.Read(c, binary.LittleEndian, answer)
		if err != nil {
			Crash("Unable to read server protocol response:", err.Error())
		}
		if answer {
			logInitialize()
			log.Println("=================================================")
			log.Println("PROTOCOL UPDATED! PLEASE UPDATE THIS APPLICATION!")
			log.Println("Download at: https://hexawolf.me/things/")
			log.Println("=================================================")
			if runtime.GOOS == "windows" {
				fmt.Println("You may close this window or it will be closed in 10 minutes.")
			} else {
				fmt.Println("Press ctrl+c to close this application. It will be closed in 10 minutes.")
			}
			time.Sleep(time.Minute * 10)
			return
		}
	}

	// Generate new UUID/load saved UUID.
	uuid, err := UUID()
	if err != nil {
		Crash("Error while loading UUID:", err.Error())
	}
	// Send it.
	log.Println("Sending UUID...")
	_, err = c.Write(uuid)
	if err != nil {
		Crash("Unable to send UUID", err.Error())
	}

	// Load hardcoded key.
	LoadKeys()
	// Read & verify signature for UUID.
	log.Println("Reading signature...")
	uuidSig := [112]byte{}
	c.Read(uuidSig[:])
	valid := Verify(uuid, uuidSig)
	if !valid {
		Crash("Invalid UUID signature received", err.Error())
	}
	log.Println("Valid UUID signature received.")

	// Send hardware information if necessary.
	shouldSend := false
	err = binary.Read(c, binary.LittleEndian, &shouldSend)
	if err != nil {
		Crash("Unable to read metrics byte from stream:", err)
	}
	if shouldSend {
		log.Print("Sending HW info... ")
		err = WriteHWInfo(c)
		if err != nil {
			Crash("Unable to send HWInfo:", err.Error())
		}
		log.Println("Sent!")
	}

	// Collect hashes of files in config/ and mods/ and send them.
	list, err := collectHashList()
	if err != nil {
		Crash("Unable to create hash list of files:", err.Error())
	}
	log.Println("Sending information about", len(list), "files...")
	resp, err := WriteHashList(list, c)
	if err != nil {
		Crash("Unable to send information about files:", err.Error())
	}

	// Apply "changes" requested by server - delete excess files.
	for k, v := range resp {
		if !v {
			if filepath.Dir(k) == "config" {
				log.Println(k, "- IGNORED")
			} else {
				log.Println(k, "- DELETE")
				os.Remove(k)
			}
		} else {
			log.Println(k, "- OK")
		}
	}

	// Apply "changes" request by server - download new files.
	for {
		log.Println("Receiving packets...")
		p, err := ReadPacket(c)
		if err != nil {
			if err == io.EOF {
				log.Println("Connection closed.")
				return
			}
			Crash("Error while receiving delta:", err.Error())
		}
		realSum := sha256.Sum256(p.Blob)
		log.Println("Received file", p.FilePath, "("+hex.EncodeToString(realSum[:])+")")

		if !p.Verify() {
			Crash("Signature check - FAILED!")
		}
		log.Println("Signature check - OK.")

		// Ensure all directories exist.
		err = os.MkdirAll(filepath.Dir(p.FilePath), 0770)
		if err != nil {
			Crash("Error while creating directories:", err.Error())
		}

		err = ioutil.WriteFile(p.FilePath, p.Blob, 0660)
		if err != nil {
			Crash("Error writing file blob:", err.Error())
		}
	}

	if runtime.GOOS == "windows" {
		time.Sleep(time.Second * 4)
	}
}
