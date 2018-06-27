package main

import (
	"os"
	"bufio"
	"strings"
)

var machinesFile *os.File

// Check if machine was already logged
func searchForMachine(id string) bool {
	machinesFile.Seek(0, 0)
	scanner := bufio.NewScanner(machinesFile)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), id) {
			return true
		}
	}
	return false
}

func writeMachine(id string, info []byte) {
	machinesFile.Write([]byte(id))
	machinesFile.Write([]byte(":"))
	machinesFile.Write(info)
	machinesFile.Write([]byte("\n"))
	machinesFile.Sync()
}
