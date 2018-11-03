// service.go - handles multiple connection into a queue properly
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
	"log"
	"net"
	"sync"
	"time"
)

type Service struct {
	quit chan bool
	wg   *sync.WaitGroup
}

func NewService() *Service {
	s := &Service{
		quit: make(chan bool),
		wg:   &sync.WaitGroup{},
	}
	return s
}

// Serve connections and spawn a goroutine to serve each one. Stop listening
// if anything is received on the service's channel.
func (s *Service) Serve(listener *net.TCPListener) {
	for {
		listener.SetDeadline(time.Now().Add(time.Second * 300))
		conn, err := listener.AcceptTCP()
		if nil != err {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			log.Println(err)
		}
		log.Println("Serving", conn.RemoteAddr())
		s.wg.Add(1)
		secureConn := tls.Server(conn, &tlsConfig)
		go s.serve(secureConn)
	}
}

// Stop the service by closing the service's channel. Block until the service
// is really stopped.
func (s *Service) Stop() {
	close(s.quit)
	s.wg.Wait()
}
