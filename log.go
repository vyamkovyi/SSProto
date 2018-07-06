package main

import (
	"log"
	"os"
	"io"
	"fmt"
	"bufio"
)

func Crash(data ...interface{}) {
	fmt.Println()
	fmt.Println("=============================================================")
	fmt.Println("\tCRASH OCCURED!")
	fmt.Println("Please contact with administrator and send ss-error.log file!")
	fmt.Println("=============================================================")
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	logFile, err := os.OpenFile("ss-error.log",
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
	if err != nil {
		fmt.Println("Looks like you don't have write access.")
		fmt.Println("You might want to run this application as an administator.")
		log.Println(err)
		log.Println("Crash cause:", data)
	} else {
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multiWriter)
		log.Println(data...)
	}
	fmt.Println("Press any key to exit.")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	os.Exit(1)
}
