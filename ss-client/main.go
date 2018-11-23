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
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/inconshreveable/go-update"
)

// SSProtoVersion is a protocol version. Used to determine if we need to update this application.
const SSProtoVersion uint8 = 2

// These variables are set by build.sh
var targetHost string
var buildStamp string

var noLaunch = false
var forceCurrent = false
var installDirectory string

// launchClient tries to launch client startup script distributed with Hexamine client.
// Notice for future generations: you likely want to get rid of this if you want reuse SSProto
// as this is purely Hexamine-specific code.
func launchClient() {
	if noLaunch {
		return
	}
	var com *exec.Cmd
	if runtime.GOOS == "windows" {
		com = exec.Command("Launch.bat")
	} else {
		os.Chmod("Launch.sh", 0775)
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

// Crash function crashes the application saving data to the ss-error.log file
func Crash(data ...interface{}) {
	fmt.Println()
	fmt.Println("=============================================================")
	fmt.Println("\tCRASH OCCURRED!")
	fmt.Println("Please contact with administrator and send ss-error.log file!")
	fmt.Println("=============================================================")
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	logFile, err := os.OpenFile("ss-error.log",
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
	if err != nil {
		fmt.Println("Looks like you don't have write access.")
		if runtime.GOOS == "windows" {
			fmt.Println("Minecraft isn't really ought to be installed in Program Files.")
		}
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

func exePath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	ex, err = filepath.Abs(ex)
	if err != nil {
		return "", err
	}
	return ex, nil
}

func handleArgs() {
	if containsString(os.Args, "--help") {
		fmt.Println("Usage:")
		fmt.Println("--only-launch \t- Do not perform any updates, just launch the game.")
		fmt.Println("--install-dir \"path\" \t- directory to install client.")
		fmt.Println("--no-launch \t- Do not launch client after installation.")
		fmt.Println("--copyright \t- License and copyright.")
		fmt.Println("--help \t\t- this.")
		os.Exit(0)
	}

	if containsString(os.Args, "--legal") {
		fmt.Println("This application uses MIT license.")
		fmt.Println(`
Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.`)
		return
	}

	if containsString(os.Args, "--only-launch") {
		launchClient()
		return
	}

	if containsString(os.Args, "--no-launch") {
		noLaunch = true
	}

	if containsString(os.Args, "--install-dir") {
		index := posString(os.Args, "--install-dir") + 1
		if len(os.Args) < index {
			fmt.Println()
			fmt.Println("Invalid usage!")
			os.Exit(1)
		}
		installDirectory = os.Args[index]
	}
}

// prepareInstallDir selects install directory and chdir's into it.
func prepareInstallDir() error {
	if installDirectory == "" {
		if runtime.GOOS == "windows" {
			installDirectory = os.Getenv("AppData") + "\\.hexamine\\"
		} else {
			installDirectory = os.Getenv("HOME") + "/.hexamine/"
		}
	}
	fmt.Println("Installation directory is", installDirectory)
	if err := os.MkdirAll(installDirectory, os.ModePerm); err != nil {
		return err
	}
	if err := os.Chdir(installDirectory); err != nil {
		return err
	}
	if err := os.MkdirAll("mods", os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll("config", os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll("versions", os.ModePerm); err != nil {
		return err
	}

	return nil
}

func runSelfupdate() {
	shouldRestart, err := downloadLatestClient()
	if err != nil {
		Crash("downloadLatestClient", err)
	}

	if shouldRestart {
		// Get full path to updater executable
		executable, err := exePath()
		if err != nil {
			Crash(err)
		}

		// Run new instance of this application
		cmd := exec.Command(executable)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			Crash(err)
		}

		os.Exit(0)
	}
}

// downloadLatestClient downloads latest Updater binary from website, replaces
// current binary with it.
//
// If this function returns true - self-update is performed and program should
// be restarted, otherwise if err=nil execution can be continued as is.
func downloadLatestClient() (bool, error) {
	var filename string
	if runtime.GOOS == "windows" {
		filename = "Updater.exe"
	} else if runtime.GOOS == "darwin" {
		filename = "Updater-mac"
	} else {
		filename = "Updater"
	}

	cl := http.Client{}

	fmt.Println("Checking for launcher updates...")

	req, err := http.NewRequest("GET", "https://"+strings.Split(targetHost, ":")[0]+"/projects/hexamine/"+filename, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("If-Modified-Since", buildStamp)

	resp, err := cl.Do(req)
	if err != nil {
		return false, err
	}
	if resp.StatusCode == http.StatusNotModified {
		// it's not necesary to perform update.
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, errors.New("HTTP " + resp.Status)
	}

	fmt.Println("Downloaded new version!")

	// Download new version and replace updater file
	fmt.Println("Applying update and starting updater again...")
	err = update.Apply(resp.Body, update.Options{})
	if err != nil {
		return false, err
	}
	resp.Body.Close()

	return true, nil
}

func savePacket(p *Packet) error {
	// Ensure all directories exist.
	err := os.MkdirAll(filepath.Dir(p.FilePath), 0775)
	if err != nil {
		return err
	}

	f, err := os.Create(p.FilePath + ".new")
	if err != nil {
		return err
	}

	err = copyWithProgress(p.FilePath, p.Size, p.Blob, f)
	if err != nil {
		f.Close()
		os.Remove(p.FilePath + ".new")
		return err
	}

	f.Close()

	err = os.Rename(p.FilePath+".new", p.FilePath)
	if err != nil {
		return err
	}

	return nil
}

func removeExcessFiles(c *tls.Conn) error {
	list, err := collectHashList()
	if err != nil {
		return err
	}
	fmt.Println("Sending information about", len(list), "files...")

	// Apply "changes" requested by server - delete excess files.
	orderedList := make([]string, 0, len(list))
	for k, v := range list {
		err := SendHashListEntry(c, k, v)
		if err != nil {
			return err
		}
		orderedList = append(orderedList, k)
	}
	for _, path := range orderedList {
		resp := true
		err := binary.Read(c, binary.LittleEndian, &resp)
		if err != nil {
			return err
		}

		if !resp && filepath.Dir(path) == "mods" {
			fmt.Println("Removing", path)
			if err := os.Remove(path); err != nil {
				fmt.Printf("Failed to remove %v: %v\n", path, err)
			}
		}
	}
	return nil
}

// main ✨✨✨
func main() {
	fmt.Println("SSProto updater")
	fmt.Println("Copyright (C) Hexawolf 2018")

	handleArgs()

	fmt.Println("SSProto version:", SSProtoVersion)
	fmt.Println("Build timestamp:", buildStamp)

	runSelfupdate()

	c, err := tls.Dial("tcp", targetHost, &conf)
	if err != nil {
		fmt.Println("Unable to connect the update server.")
		fmt.Println("If you really want to start Hexamine client without updating, " +
			"run updater with --only-launch flag.")
		Crash("tls.Dial", err)
	}
	defer c.Close()

	defer time.Sleep(time.Second * 5)

	// Setting up directory
	if err := prepareInstallDir(); err != nil {
		Crash("prepareInstallDir", err)
	}

	// Check protocol version
	{
		err = binary.Write(c, binary.LittleEndian, SSProtoVersion)
		if err != nil {
			Crash("Unable to send SSProto version:", err.Error())
		}
		var pv uint8
		err = binary.Read(c, binary.LittleEndian, &pv)
		if err != nil {
			Crash("Unable to read server protocol response:", err.Error())
		}
		fmt.Println("Server protocol version:", pv)
		if pv != SSProtoVersion {
			runSelfupdate()
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

	connectionAccepted := false
	err = binary.Read(c, binary.LittleEndian, &connectionAccepted)
	if err != nil {
		Crash("Unable to read connection status byte from stream:", err)
	}
	if connectionAccepted {
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
	fmt.Println("Hashing all files...")
	// TODO: This thing can be merged together with code below to increase performance.
	// E.g. pipeining, send file info right after hashing it.
	if err := removeExcessFiles(c); err != nil {
		Crash(err)
	}

	zeroes := [32]byte{}
	_, err = c.Write(zeroes[:])
	if err != nil {
		Crash("Failed to send hashlist terminator:", err)
	}

	// Apply "changes" request by server - download new files.
	fmt.Println("Listening for packets...")
	for {
		p, err := ReadPacket(c)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed.")
				launchClient()
				return
			}
			Crash("Error while receiving delta:", err.Error())
		}

		if err := savePacket(p); err != nil {
			Crash("savePacket", err)
		}
	}
}
