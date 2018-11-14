// main.go - listening to connections and handling them in a separate goroutine
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
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

// SSProtoVersion is a protocol version. Used to determine if clients need update.
const SSProtoVersion uint8 = 2

var tlsConfig tls.Config

var serverConfig Config

func main() {
	// Rotate logs and set up logging to both file and stdout
	// See logging.go
	LogInitialize()
	log.Println("SSProto version", SSProtoVersion)
	log.Println("Copyright (C) Hexawolf  2018")
	var err error

	// Loading server config
	err = serverConfig.LoadConfig("ssserver.toml")
	if err != nil {
		log.Panicln("Failed to read server config:", err)
	}

	// Initialize TLS
	var cert tls.Certificate
	cert, err = tls.LoadX509KeyPair(serverConfig.Certificate, serverConfig.Key)
	if err != nil {
		log.Panicln("Failed to initialize TLS:", err)
	}
	tlsConfig = tls.Config{
		Certificates:       []tls.Certificate{cert},
		ServerName:         serverConfig.ServerName,
		InsecureSkipVerify: true,
	}

	// Prepares served files list
	// lister.go
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Panicln("Failed to initialize fsnotify:", err)
	}
	ListFiles()
	go handleFSEvents()

	defer logFile.Close()

	laddr, err := net.ResolveTCPAddr("tcp", serverConfig.Address)
	if err != nil {
		log.Panicln("Error listening:", err)
	}

	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Panicln("Error listening:", err)
	}
	// Close the listener when the application closes.
	defer l.Close()
	log.Println("Listening on", serverConfig.Address)

	// Start network message processing service
	service := NewService()
	go service.Serve(l)

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-c
	fmt.Println()
	log.Println("Signal caught, waiting for connections to close and exiting...")
	service.Stop()
}
