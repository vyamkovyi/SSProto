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
	"regexp"

	"crypto/tls"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var excludedGlob = []string{
	"/?ignored_*",
	"assets",
	"screenshots",
	"saves",
	"library",
}

func shouldExclude(path string) bool {
	for _, pattern := range excludedGlob {
		if match, _ := regexp.MatchString(pattern, filepath.ToSlash(path)); match {
			return true
		}
	}
	return false
}

func collectRecurse(root string) ([]string, error) {
	var res []string
	walkfn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if shouldExclude(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldExclude(path) {
			return nil
		}

		res = append(res, path)
		return nil
	}
	err := filepath.Walk(root, walkfn)
	return res, err
}

func collectHashList() (map[string][]byte, error) {
	res := make(map[string][]byte)

	list, err := collectRecurse(".")
	if err != nil {
		return nil, err
	}

	authlib := "libraries/com/mojang/authlib/1.5.25/authlib-1.5.25.jar"
	if fileExists(authlib) {
		list = append(list, filepath.ToSlash(authlib))
	}

	for _, path := range list {
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

func askForConfirmation() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		Crash(err)
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

func launchClient() {
	var com *exec.Cmd = nil
	if runtime.GOOS == "windows" {
		com = exec.Command("Launch.bat")
	} else if runtime.GOOS == "linux" {
		os.Chmod("Launch.sh", 0770)
		com = exec.Command("./Launch.sh")
	}
	if com != nil {
		com.Run()
	}
}

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

		if !(checkSecond && checkFirst && checkThird) {
			return true
		}
	}
	return false
}

const SSProtoVersion uint8 = 1
const TargetHost = "doggoat.de"

func main() {
	fmt.Println("SSProto version:", SSProtoVersion)
	fmt.Println("Copyright (C) Hexawolf, foxcpp 2018")
	// Load hardcoded key.
	LoadKeys()

	c, err := tls.Dial("tcp", TargetHost+":48879", &conf)
	if err != nil {
		Crash("Unable to connect to the server:", err.Error())
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
			fmt.Println("Download at https://hexawolf.me/hexamine/" + filename)
			fmt.Println("=================================================")
			if runtime.GOOS == "windows" {
				fmt.Println("You may now close this window.")
			} else {
				fmt.Println("Press ctrl+c to close this application.")
			}
			time.Sleep(time.Minute * 10)
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
