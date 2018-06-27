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

	"github.com/twstrike/ed448"
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

const SSProtoVersion = 1

func main() {
	log.Println("SSProto version:", SSProtoVersion)
	log.Println("Copyright (C) Hexawolf, foxcpp 2018")

	c, err := net.Dial("tcp", "monad:48879")
	if err != nil {
		log.Println("IO:", err)
		return
	}
	defer c.Close()

	// Generate new UUID/load saved UUID.
	uuid, err := UUID()
	if err != nil {
		log.Println("UUID get:", err)
		return
	}
	// Send it.
	log.Println("Sending UUID...")
	_, err = c.Write(uuid)
	if err != nil {
		log.Println("UUID send:", err)
		return
	}

	// Load hardcoded key.
	key := loadKeys()
	// Read & verify signature for UUID.
	log.Println("Reading signature...")
	uuidSig := [112]byte{}
	c.Read(uuidSig[:])
	curve := ed448.NewDecafCurve()
	valid, err := curve.Verify(uuidSig, uuid, key)
	if err != nil {
		log.Println(err)
		return
	}
	if !valid {
		log.Println("Invalid UUID sig")
		return
	} else {
		log.Println("Valid UUID signature received.")
	}

	// Send hardware information if necessary.
	shouldSend := false
	err = binary.Read(c, binary.LittleEndian, &shouldSend)
	if err != nil {
		log.Println("IO:", err)
		return
	}
	if shouldSend {
		log.Print("Sending HW info... ")
		err = WriteHWInfo(c)
		if err != nil {
			log.Println("HWInfo send:", err)
			return
		}
		log.Println("Sent!")
	}

	// Collect hashes of files in config/ and mods/ and send them.
	list, err := collectHashList()
	if err != nil {
		log.Println("Hash list collect:", err)
		return
	}
	log.Println("Sending information about", len(list), "files...")
	resp, err := WriteHashList(list, c)
	if err != nil {
		log.Println("Hash list send:", err)
		return
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
				log.Println("Finishing")
				return
			}
			log.Println("IO:", err)
			return
		}
		realSum := sha256.Sum256(p.Blob)
		log.Println("Received file", p.FilePath, "("+hex.EncodeToString(realSum[:])+")")

		if !p.Verify(key) {
			log.Println("Invalid signature received")
			return
		}
		log.Println("Packet signature - OK.")

		// Ensure all directories exist.
		err = os.MkdirAll(filepath.Dir(p.FilePath), 0770)
		if err != nil {
			log.Println("FS:", err)
			return
		}

		err = ioutil.WriteFile(p.FilePath, p.Blob, 0660)
		if err != nil {
			log.Println("FS:", err)
			return
		}
	}
}
