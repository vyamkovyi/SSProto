package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"crypto/tls"
	"runtime"
	"time"
	"bufio"
)

const SSProtoVersion uint8 = 1

// This variable is set by build.sh
var targetHost string

func main() {
	fmt.Println("SSProto version:", SSProtoVersion)
	fmt.Println("Copyright (C) Hexawolf, foxcpp 2018")
	// Load hardcoded key.
	LoadKeys()

	c, err := tls.Dial("tcp", targetHost, &conf)
	if err != nil {
		com := "./Launch.sh"
		if runtime.GOOS == "windows" {
			com = "Launch.bat"
		}
		fmt.Println("Unable to connect the update server.")
		fmt.Println("If you really want to start Hexamine client without updating, run", com)
		Crash(err.Error())
	}
	defer c.Close()
	defer time.Sleep(time.Second * 5)

	if checkDir() {
		fmt.Println()
		fmt.Println("=================================================================")
		fmt.Println("! Make sure this application was launched under a new directory !")
		fmt.Println("=================================================================")
		fmt.Println("The updater will download files right into current directory.")
		fmt.Println("However, it does not looks like an empty directory or existing client. You probably don't want to download files here.")
		fmt.Print("Do you want to proceed? (y/n): ")
		for !askForConfirmation() {
			fmt.Println("Exiting.")
			return
		}
	}

	os.MkdirAll("mods", 0770)
	os.MkdirAll("config", 0770)
	os.MkdirAll("versions", 0770)

	// Send protocol version and get answer whether we must ask user for update
	{
		err = binary.Write(c, binary.LittleEndian, SSProtoVersion)
		if err != nil {
			Crash("Unable to send SSProto version:", err.Error())
		}
		var answer bool
		err = binary.Read(c, binary.LittleEndian, answer)
		if err != nil {
			Crash("Unable to read server protocol response:", err.Error())
		}
		if answer {
			c.Close()
			filename := ""
			if runtime.GOOS == "windows" {
				filename = "Updater.exe"
			} else if runtime.GOOS == "linux" {
				filename = "Updater"
			}
			fmt.Println()
			fmt.Println("=================================================")
			fmt.Println("PROTOCOL UPDATED! PLEASE UPDATE THIS APPLICATION!")
			fmt.Println("=================================================")
			fmt.Println("Download at https://hexawolf.me/hexamine/" + filename)
			fmt.Println()
			fmt.Println("Press enter to exit.")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
			os.Exit(0)
		}
	}

	// Generate new UUID/load saved UUID.
	uuid, err := UUID()
	if err != nil {
		Crash("Error while loading UUID:", err.Error())
	}
	fmt.Println("Our UUID:", base64.StdEncoding.EncodeToString(uuid))
	// Send it.
	fmt.Println("Sending UUID...")
	_, err = c.Write(uuid)
	if err != nil {
		Crash("Unable to send UUID", err.Error())
	}

	// Read & verify signature for UUID.
	fmt.Println("Reading signature...")
	uuidSig := [112]byte{}
	c.Read(uuidSig[:])
	valid := Verify(uuid, uuidSig)
	if !valid {
		Crash("Invalid UUID signature received")
	}
	fmt.Println("Valid UUID signature received.")

	// Send hardware information if necessary.
	shouldSend := false
	err = binary.Read(c, binary.LittleEndian, &shouldSend)
	if err != nil {
		Crash("Unable to read metrics byte from stream:", err)
	}
	if shouldSend {
		fmt.Print("Sending HW info... ")
		err = WriteHWInfo(c)
		if err != nil {
			Crash("Unable to send HWInfo:", err.Error())
		}
		fmt.Println("Sent!")
	} else {
		fmt.Println("Server rejected download request. " +
			"Simply launching client for now.")
		launchClient()
		return
	}

	// Collect hashes of files in config/ and mods/ and send them.
	list, err := collectHashList()
	if err != nil {
		Crash("Unable to create hash list of files:", err.Error())
	}
	fmt.Println("Sending information about", len(list), "files...")

	// Apply "changes" requested by server - delete excess files.
	for k, v := range list {
		fmt.Print(k)
		resp, err := SendHashListEntry(c, k, v)
		if err != nil {
			fmt.Println(" - FAIL:", err)
			Crash("Failed to send info about", k+":", err)
		}

		if !resp {
			if filepath.Dir(k) != "mods" {
				fmt.Println(" - IGNORED")
			} else {
				fmt.Println(" - DELETE")
				os.Remove(k)
			}
		} else {
			fmt.Println(" - OK")
		}
	}
	err = FinishHashList(c)
	if err != nil {
		Crash("Failed to send hashlist terminator:", err)
	}

	// Apply "changes" request by server - download new files.
	for {
		fmt.Println("Receiving packets...")
		p, err := ReadPacket(c)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed.")
				launchClient()
				return
			}
			Crash("Error while receiving delta:", err.Error())
		}
		realSum := sha256.Sum256(p.Blob)
		fmt.Println("Received file", p.FilePath,
			"("+hex.EncodeToString(realSum[:])+")")

		if !p.Verify() {
			Crash("Signature check - FAILED!")
		}
		fmt.Println("Signature check - OK.")

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
}
