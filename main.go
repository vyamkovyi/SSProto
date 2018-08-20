// main.go - wraps everything up and performs SSProto magic ✨
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
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"bufio"
	"crypto/tls"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// SSProto protocol version. Used to determine if we need to update our updater.
const SSProtoVersion uint8 = 1

// This variable is set by build.sh
var targetHost string

// launchClient tries to launch client startup script distributed with Hexamine client.
// Notice for future generations: you likely want to get rid of this if you want reuse SSProto
// as this is purely Hexamine-specific code.
func launchClient() {
	var com *exec.Cmd = nil
	if runtime.GOOS == "windows" {
		com = exec.Command("Launch.bat")
	} else {
		os.Chmod("Launch.sh", 0770)
		com = exec.Command("./Launch.sh")
	}
	err := com.Run()
	if err != nil {
		fmt.Println()
		fmt.Println("==================================")
		fmt.Println("Client was installed successfully!")
		fmt.Println("==================================")
		fmt.Println("However, we were unable to start TLauncher.")
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println("!MAKE SURE JAVA IS INSTALLED AND RUN UPDATER AGAIN!")
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println("Press enter to exit.")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}

// checkDir tries to do esoteric scanning to see if current directory suitable for installing client
func checkDir() bool {
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

		return !(checkSecond && checkFirst && checkThird)
	}
	return true
}

// Crash function crashes the application saving data to the ss-error.log file
func Crash(data ...interface{}) {
	fmt.Println()
	fmt.Println("=============================================================")
	fmt.Println("\tCRASH OCCURRED!")
	fmt.Println("Please contact with administrator and send ss-error.log file!")
	fmt.Println("=============================================================")
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	logFile, err := os.OpenFile("ss-error.log",
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
	if err != nil {
		fmt.Println("Looks like you don't have write access.")
		if runtime.GOOS == "windows" {
			fmt.Println("Minecraft isn't really ought to be installed in Program Files.")
		}
		fmt.Println("You might want to run this application as administator if you don't really care about" +
			"security. Alternatively, create directory in your user's home directory and install client there.")
		log.Println(err)
		log.Println("Crash cause:", data)
	} else {
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multiWriter)
		log.Println(data...)
	}
	fmt.Println("Press enter to exit.")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	os.Exit(1)
}

// main ✨✨✨
func main() {
	fmt.Println("SSProto version:", SSProtoVersion)
	fmt.Println("Copyright (C) Hexawolf 2018")
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

	fmt.Println()
	installDirectory := os.Getenv("HOME") + "/.hexamine/"
	if runtime.GOOS == "windows" {
		installDirectory = os.Getenv("AppData") + "\\.hexamine\\"
	}
	currentDirStatus := checkDir()
	if currentDirStatus {
		fmt.Println("Default directory for installation is \"" + installDirectory + "\".")
		fmt.Print("Do you want to use current directory instead? (y/n): ")
		if askForConfirmation() {
			installDirectory = "./"
			if runtime.GOOS == "windows" {
				installDirectory = ".\\"
			}
		} else {
			os.MkdirAll(installDirectory, 0770)
		}
		os.MkdirAll(installDirectory+"mods", 0770)
		os.MkdirAll(installDirectory+"config", 0770)
		os.MkdirAll(installDirectory+"versions", 0770)
	} else {
		installDirectory = "."
	}
	os.Chdir(installDirectory)

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
