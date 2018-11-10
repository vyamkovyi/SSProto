// logging.go - log rotation and initialization
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
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Rotate function performs logs rotation in specified directory. It checks existence of prefix.N.suffix files and
// renames prefix.suffix file to prefix.N+1.suffix file. That is, log with largest N is the latest log.
func Rotate(prefix string, suffix string, directory string) {
	prefix = prefix + "."
	suffix = "log"

	// List all files in logs/ directory or create a new directory
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(directory, 0740)
			return
		}
		log.Fatal(err)
	}

	// Collect list of real log files
	var logs []string
	for _, i := range files {
		if strings.Contains(i.Name(), prefix) &&
			strings.Contains(i.Name(), suffix) {
			logs = append(logs, i.Name())
		}
	}

	// Rename existing log
	sort.Sort(sort.StringSlice(logs))
	num := 0
	for _, i := range logs {
		arr := strings.Split(i, ".")
		num, err = strconv.Atoi(arr[len(arr)-2])
		if err == nil {
			break
		}
	}
	os.Rename(directory+prefix+suffix, directory+prefix+strconv.Itoa(num+1)+"."+suffix)
}

var logFile *os.File

// LogInitialize sets up logging with date and time in UTC format to both file and stdout.
func LogInitialize() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	Rotate("sss", "log", "logs/")
	var err error
	logFile, err = os.OpenFile("logs/sss.log",
		os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		log.Panicln(err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
}
