package main

import (
	"io"
	"log"
	"os"
)

func logInitialize() {
	logFile, err := os.OpenFile("ss-error.log",
		os.O_CREATE|os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		log.Panicln(err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
}

func Crash(data ...interface{}) {
	logInitialize()
	log.Fatalln(data...)
}
