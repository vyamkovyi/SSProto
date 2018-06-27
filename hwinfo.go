package main

import (
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"unicode"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
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

type CPUStat struct {
	CPU       int32   `json:"cpu"`
	VendorID  string  `json:"vendorId"`
	Family    string  `json:"family"`
	Model     string  `json:"model"`
	Cores     int32   `json:"cores"`
	ModelName string  `json:"modelName"`
	Mhz       float64 `json:"mhz"`
	CacheSize int32   `json:"cacheSize"`
}

func getCPUInfo() CPUStat {
	stat, err := cpu.Info()
	var stat2 CPUStat
	if err != nil {
		stat2.CacheSize = stat[0].CacheSize
		stat2.Cores = stat[0].Cores
		stat2.CPU = stat[0].CPU
		stat2.Family = stat[0].Family
		stat2.Mhz = stat[0].Mhz
		stat2.Model = stat[0].Model
		stat2.ModelName = stat[0].ModelName
		stat2.VendorID = stat[0].VendorID
	}
	return stat2
}

func getGPUInfo() string {
	if runtime.GOOS == "windows" {
		Info := exec.Command("cmd", "/C", "wmic path win32_VideoController get name")
		Info.SysProcAttr = &syscall.SysProcAttr{}
		History, _ := Info.Output()
		return strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, strings.Replace(string(History), "Name", "", -1))
	}
	return ""
}

type MachineInfo struct {
	Mem VirtualMemory
	Cpu CPUStat
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
