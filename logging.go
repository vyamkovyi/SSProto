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
	"sync"
	"bufio"
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
		os.O_CREATE | os.O_RDWR, 0660)
	if err != nil {
		log.Panicln(err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
}

var mut = &sync.Mutex{}

// Check if machine was already logged
func machineExists(id string) bool {
	mut.Lock()
	logFile.Seek(0, 0)
	defer logFile.Seek(0, 2)
	defer mut.Unlock()
	scanner := bufio.NewScanner(logFile)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), id) {
			return true
		}
	}
	return false
}
