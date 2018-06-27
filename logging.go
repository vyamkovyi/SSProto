package main

import (
	"io/ioutil"
	"os"
	"log"
	"strings"
	"sort"
	"strconv"
	"io"
)

func logRotate() {
	files, err := ioutil.ReadDir("logs/")
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir("logs/", 0740)
			return
		} else {
			log.Fatal(err)
		}
	}

	var logs []string
	for _, i := range files {
		if strings.Contains(i.Name(), "sss.") &&
			strings.Contains(i.Name(), ".log") {
				logs = append(logs, i.Name())
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(logs)))
	for _, i := range logs {
		arr :=  strings.Split(i, ".")
		num, err := strconv.Atoi(arr[len(arr)-2])
		if err != nil {
			log.Fatal(err)
		}
		os.Rename("logs/" + i, "logs/sss." + strconv.Itoa(num+1) + ".log")
	}
}

func logInitialize() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	logRotate()
	logFile, err := os.OpenFile("logs/sss.0.log",
		os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0660)
	if err != nil {
		log.Panicln(err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
}
