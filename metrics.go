package main

import (
	"os"
	"bufio"
	"strings"
)

type VirtualMemory struct {
	// Total amount of RAM on this system
	Total uint64 `json:"total"`

	// RAM available for programs to allocate
	//
	// This value is computed from the kernel specific values.
	Available uint64 `json:"available"`

	// RAM used by programs
	//
	// This value is computed from the kernel specific values.
	Used uint64 `json:"used"`

	// Percentage of RAM used by programs
	//
	// This value is computed from the kernel specific values.
	UsedPercent float64 `json:"usedPercent"`

	// This is the kernel's notion of free memory; RAM chips whose bits nobody
	// cares about the value of right now. For a human consumable number,
	// Available is what you really want.
	Free uint64 `json:"free"`
}

type CPUStat struct {
	CPU        int32    `json:"cpu"`
	VendorID   string   `json:"vendorId"`
	Family     string   `json:"family"`
	Model      string   `json:"model"`
	Cores      int32    `json:"cores"`
	ModelName  string   `json:"modelName"`
	Mhz        float64  `json:"mhz"`
	CacheSize  int32    `json:"cacheSize"`
}

type MachineInfo struct {
	mem VirtualMemory
	cpu CPUStat
	gpu string
	os string
}

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
