package main

import (
	"log"
	"net"
	"time"
	"sync"
	"crypto/tls"
)

type Service struct {
	quit chan bool
	wg *sync.WaitGroup
}

func NewService() *Service {
	s := &Service{
		quit: make(chan bool),
		wg: &sync.WaitGroup{},
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