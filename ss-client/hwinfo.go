// hwinfo.go - virtual memory statistics
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
