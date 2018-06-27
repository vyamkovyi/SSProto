package main

import (
	"net"
	"log"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const SSProtoVersion uint8 = 1
const address = "0.0.0.0:48879"

func main() {
	// Rotate logs and set up logging to both file and stdout
	// See logging.go
	LogInitialize()
	log.Println("SSProto version", SSProtoVersion)
	log.Println("Copyright (C) Hexawolf  2018")

	// See crypto.go
	if _, err := os.Stat("ss.key"); err != nil {
		MakeKeys()
	} else {
		LoadKeys()
	}

	// Prepares served files list
	// lister.go
	ListFiles()

	var err error
	// metrics.go
	machinesFile, err = os.OpenFile("machines.dat", os.O_RDWR | os.O_CREATE, 0600)
	if err != nil {
		log.Panicln("Failed to initialize storage!")
	}
	defer machinesFile.Close()

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
