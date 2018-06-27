package main

import (
	"net"
	"log"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"io/ioutil"
	"path/filepath"
	"crypto/sha256"
)

const version = 1
const address = "0.0.0.0:48879"

var filesMap map[[32]byte]string

func main() {
	// Rotate logs and log to both file and stdout
	logInitialize()
	log.Println("SS-Server",
		version, "Copyright (C) Hexawolf  2018")

	/*
	if _, err := os.Stat("sss.key"); err != nil {
		makeKeys()
	} else {
		loadKeys()
	}
	*/
	loadKeys()
	filesMap = make(map[[32]byte]string)

	var err error
	machinesFile, err = os.OpenFile("machines.dat", os.O_RDWR | os.O_CREATE, 0600)
	if err != nil {
		log.Panicln("Failed to initialize storage!")
	}
	defer machinesFile.Close()

	// Generate hash for all synchronizable files
	{
		// for mods/
		files, err := ioutil.ReadDir("mods/")
		if err != nil {
			log.Fatalln("Failed to read mods directory:", err.Error())
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}

			fullFileName := "mods/" + f.Name()
			if filepath.Ext(fullFileName) != ".jar" {
				continue
			}

			s, err := ioutil.ReadFile(fullFileName)
			if err != nil {
				log.Fatalln("Failed to read file", fullFileName, ":", err.Error())
			}
			hash := sha256.Sum256(s)

			filesMap[hash] = fullFileName
		}

		// for config/
		err = filepath.Walk("config/", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Fatalln("Failed to read config directory!", err.Error())
			}

			// TODO: skip unwanted directories and files
			if info.IsDir() {
				return nil
			}

			s, err := ioutil.ReadFile(path)
			if err != nil {
				log.Fatalln("Failed to read file", path, ":", err.Error())
			}
			hash := sha256.Sum256(s)

			filesMap[hash] = path
			return nil
		})
	}

	laddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Panicln("Error listening:", err.Error())
	}

	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Panicln("Error listening:", err.Error())
	}
	// Close the listener when the application closes.
	defer l.Close()
	log.Println("Listening on", address)

	// Start network message processing service
	service := NewService()
	go service.Serve(l)

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-c
	fmt.Println()
	log.Println("Signal caught, exiting!")
}
