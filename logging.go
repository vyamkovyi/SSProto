package main

import (
	"io/ioutil"
	"os"
	"log"
	"strings"
	"sort"
	"strconv"
	"io"
	"github.com/jasonlvhit/gocron"
)

func rotate(prefix string, suffix string, directory string) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(directory, 0740)
			return
		} else {
			log.Fatal(err)
		}
	}

	var logs []string
	for _, i := range files {
		if strings.Contains(i.Name(), prefix) &&
			strings.Contains(i.Name(), suffix) {
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
		os.Rename(directory + i, directory + prefix + strconv.Itoa(num+1) +
			suffix)
	}
}

var logFile *os.File

func LogInitialize() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	rotate("sss.", ".log", "logs/")
	var err error = nil
	logFile, err = os.OpenFile("logs/sss.0.log",
		os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0660)
	if err != nil {
		log.Panicln(err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	// Rotate logs every day
	gocron.Every(1).Day().At("00:00").Do(LogInitialize)
}
