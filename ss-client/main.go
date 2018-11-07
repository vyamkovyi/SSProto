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
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/inconshreveable/go-update"
	"io"
	"net/http"
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

// SSProtoVersion is a protocol version. Used to determine if we need to update this application.
const SSProtoVersion uint8 = 2

// This variable is set by build.sh
var targetHost string

var noLaunch = false
var forceCurrent = false
var installDirectory = "." + string(os.PathSeparator)

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
		fmt.Println("--force-current \t- Disable any directory checks and use current dir.")
		fmt.Println("--only-launch \t- Do not perform any updates, just launch the game.")
		fmt.Println("--install-dir \"path\" \t- directory to install client.")
		fmt.Println("--no-launch \t- Do not launch client after installation.")
		fmt.Println("--copyright \t- License and copyright.")
		fmt.Println("--help \t\t- this.")
		return
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

	if containsString(os.Args, "--force-current") {
		forceCurrent = true
	}

	if containsString(os.Args, "--install-dir") {
		forceCurrent = true
		index := posString(os.Args, "--install-dir") + 1
		if len(os.Args) < index {
			fmt.Println()
			fmt.Println("Invalid usage!")
			os.Exit(1)
		}
		installDirectory = os.Args[index]
		if !strings.ContainsRune(installDirectory, os.PathSeparator) {
			fmt.Println("Invalid path!")
			os.Exit(1)
		}
	}
}

// main ✨✨✨
func main() {
	fmt.Println("SSProto, protocol version:", SSProtoVersion)
	fmt.Println("Copyright (C) Hexawolf 2018")

	handleArgs()

	fmt.Println("SSProto version:", SSProtoVersion)
	// Load hardcoded key.
	LoadKeys()

	c, err := tls.Dial("tcp", targetHost, &conf)
	if err != nil {
		fmt.Println("Unable to connect the update server.")
		fmt.Println("If you really want to start Hexamine client without updating, " +
			"run updater with --only-launch flag.")
		Crash(err.Error())
	}
	defer c.Close()
	defer time.Sleep(time.Second * 5)

	// Setting up directory
	fmt.Println()
	if !forceCurrent {
		if runtime.GOOS == "windows" {
			installDirectory = os.Getenv("AppData") + "\\.hexamine\\"
		} else {
			installDirectory = os.Getenv("HOME") + "/.hexamine/"
		}
	}
	fmt.Println("Default directory for installation is", installDirectory)
	os.MkdirAll(installDirectory+"mods", 0775)
	os.MkdirAll(installDirectory+"config", 0770)
	os.MkdirAll(installDirectory+"versions", 0775)
	os.Chdir(installDirectory)

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
			c.Close()
			var filename string
			if runtime.GOOS == "windows" {
				filename = "Updater.exe"
			} else if runtime.GOOS == "darwin" {
				filename = "Updater-mac"
			} else {
				filename = "Updater"
			}
			fmt.Println("Protocol version has changed, downloading latest updater version!")
			resp, err := http.Get("https://" + strings.Split(targetHost, ":")[0] + "/projects/hexamine/" + filename)
			if err != nil {
				Crash("Protocol update failure:", err)
			}
			// Get full path to updater executable
			executable, err := exePath()
			if err != nil {
				Crash("Protocol update failure:", err)
			}
			// Download new version and replace updater file
			fmt.Println("Applying update and starting updater again...")
			err = update.Apply(resp.Body, update.Options{})
			if err != nil {
				Crash("Protocol update failure:", err)
			}
			resp.Body.Close()
			// Run new instance of this application
			err = exec.Command(executable).Run()
			if err != nil {
				fmt.Println("A fatal error occurred causing new updater to fail:", err)
				os.Exit(1)
			}
			return
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

		// Ensure all directories exist.
		err = os.MkdirAll(filepath.Dir(p.FilePath), 0775)
		if err != nil {
			Crash("Error while creating directories:", err.Error())
		}

		f, err := os.Create(p.FilePath + ".new")
		if err != nil {
			Crash("Error while opening write for writting:", err.Error())
		}

		err = copyWithProgress(p.FilePath, p.Size, p.Blob, f)
		if err != nil {
			f.Close()
			os.Remove(p.FilePath + ".new")
			Crash("Error writing file blob:", err.Error())
		}

		f.Close()

		err = os.Rename(p.FilePath+".new", p.FilePath)
		if err != nil {
			Crash("Error while rename new file:", err.Error())
		}
	}
}
