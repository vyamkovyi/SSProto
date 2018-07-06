package main

import (
	"runtime"
	"github.com/shirou/gopsutil/mem"
)

type MachineInfo struct {
	// Total amount of RAM on this system
	MemoryTotal uint64 `json:"mem_total"`

	// RAM available for programs to allocate
	//
	// This value is computed from the kernel specific values.
	MemoryAvailable uint64 `json:"mem_available"`

	// RAM used by programs
	//
	// This value is computed from the kernel specific values.
	MemoryUsed uint64 `json:"mem_used"`

	// Percentage of RAM used by programs
	//
	// This value is computed from the kernel specific values.
	MemoryUsedPercent float64 `json:"mem_usedPercent"`

	// This is the kernel's notion of free memory; RAM chips whose bits nobody
	// cares about the value of right now.
	MemoryFree uint64 `json:"mem_free"`

	// User's operating system, just GOOS variable
	OS  string `json:"os"`
}

func GetMachineInfo() MachineInfo {
	var info MachineInfo
	v, err := mem.VirtualMemory()
	if err == nil {
		info.MemoryAvailable = v.Available
		info.MemoryFree = v.Free
		info.MemoryTotal = v.Total
		info.MemoryUsed = v.Used
		info.MemoryUsedPercent = v.UsedPercent
	}
	info.OS = runtime.GOOS
	return info
}
