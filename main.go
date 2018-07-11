package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"crypto/tls"
)

const SSProtoVersion uint8 = 1
// This variable is set by release.sh
const address = "0.0.0.0:48879"

var tlsConfig tls.Config

func main() {
	// Rotate logs and set up logging to both file and stdout
	// See logging.go
	LogInitialize()
	log.Println("SSProto version", SSProtoVersion)
	log.Println("Copyright (C) Hexawolf  2018")
	var err error

	// See crypto.go
	if _, err := os.Stat("ss.key"); err != nil {
		MakeKeys()
	} else {
		LoadKeys()
	}

	// Initialize TLS
	var cert tls.Certificate
	cert, err = tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Panicln("Failed to initialize TLS:", err)
	}
	tlsConfig = tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName: "doggoat.de",
		InsecureSkipVerify: true,
	}

	// Prepares served files list
	// lister.go
	ListFiles()

	defer logFile.Close()

	laddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Panicln("Error listening:", err)
	}

	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Panicln("Error listening:", err)
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
