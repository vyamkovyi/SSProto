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
	"os/exec"
	"strings"
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

// posString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

// containsString returns true iff slice contains element
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}

// askForConfirmation uses Scanln to parse user input. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user. Typically, you should use fmt to print out a question
// before calling askForConfirmation. E.g. fmt.Println("WARNING: Are you sure? (yes/no)")
func askForConfirmation() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}
	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if containsString(okayResponses, response) {
		return true
	} else if containsString(nokayResponses, response) {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return askForConfirmation()
	}
}

const SSProtoVersion uint8 = 1

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	log.Println("SSProto version:", SSProtoVersion)
	log.Println("Copyright (C) Hexawolf, foxcpp 2018")

	{
		files, err := ioutil.ReadDir(".")
		if err != nil {
			Crash("Unable to read current directory:", err.Error())
		}
		if len(files) > 1 {
			checkFirst := false
			checkSecond := false
			checkThird := false
			for _, v := range files {
				if strings.Contains(v.Name(), "versions") {
					checkFirst = true
				} else if strings.Contains(v.Name(), "mods") {
					checkSecond = true
				} else if strings.Contains(v.Name(), "config") {
					checkThird = true
				}
			}

			if !(checkSecond && checkFirst && checkThird) {
				fmt.Println()
				fmt.Println("=================================================================")
				fmt.Println("! Make sure this application was launched under a new directory !")
				fmt.Println("=================================================================")
				fmt.Println("The updater will download files right into current directory. " +
					"However, it does not looks like an empty directory or existing client. " +
					"You probably don't want to download files here.")
				fmt.Print("Do you want to proceed? (y/n): ")
				for !askForConfirmation() {
					fmt.Println("Exiting.")
					return
				}
 			}
		}
	}

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
			fmt.Println()
			fmt.Println("=================================================")
			fmt.Println("PROTOCOL UPDATED! PLEASE UPDATE THIS APPLICATION!")
			fmt.Println("Download at: https://hexawolf.me/things/")
			fmt.Println("=================================================")
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

	var com *exec.Cmd = nil
	if runtime.GOOS == "windows" {
		com = exec.Command("java -jar libraries\\TLauncher.jar " +
			"--directory . --settings config\\tlauncher.cfg --profiles " +
			"config\\tlauncher_profiles.json --version \"Hexamine\"")
	} else if runtime.GOOS == "linux" {
		os.Chmod("Launch.sh", 0770)
		com = exec.Command("java -jar ./libraries/TLauncher.jar " +
			"--directory ./ --settings ./config/tlauncher.cfg --profiles " +
			"./config/tlauncher_profiles.json --version \"Hexamine\"")
	}
	if com != nil {
		com.Run()
	}
}
