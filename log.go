package main

import (
	"log"
	"os"
	"io"
	"fmt"
	"bufio"
	"runtime"
)

func Crash(data ...interface{}) {
	fmt.Println()
	fmt.Println("=============================================================")
	fmt.Println("\tCRASH OCCURRED!")
	fmt.Println("Please contact with administrator and send ss-error.log file!")
	fmt.Println("=============================================================")
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	logFile, err := os.OpenFile("ss-error.log",
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
	if err != nil {
		fmt.Println("Looks like you don't have write access.")
		if runtime.GOOS == "windows" {
			fmt.Println("Minecraft isn't really ought to be installed in Program Files.")
		}
		fmt.Println("You might want to run this application as administator if you don't really care about" +
			"security. Alternatively, create directory in your user's home directory and install client there.")
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
