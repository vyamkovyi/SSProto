package main

import (
	"os/exec"
	"runtime"
	"syscall"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
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

func getMemoryInfo() VirtualMemory {
	v, err := mem.VirtualMemory()
	var v2 VirtualMemory
	if err == nil {
		v2.Available = v.Available
		v2.Free = v.Free
		v2.Total = v.Total
		v2.Used = v.Used
		v2.UsedPercent = v.UsedPercent
	}
	return v2
}

func getCPUInfo() []cpu.InfoStat {
	stat, _ := cpu.Info()
	return stat
}

func getGPUInfo() string {
	if runtime.GOOS == "windows" {
		Info := exec.Command("cmd", "/C",
			"wmic path win32_VideoController get name")
		Info.SysProcAttr = &syscall.SysProcAttr{}
		History, _ := Info.Output()
		outputString := strings.TrimPrefix(string(History), "Name")
		return strings.Trim(outputString, " \r\n")
	}
	return ""
}

type MachineInfo struct {
	Mem VirtualMemory
	Cpu []cpu.InfoStat
	Gpu string
	OS  string
}

func GetMachineInfo() MachineInfo {
	var info MachineInfo
	info.Cpu = getCPUInfo()
	info.Gpu = getGPUInfo()
	info.Mem = getMemoryInfo()
	info.OS = runtime.GOOS
	return info
}
