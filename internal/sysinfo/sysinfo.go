// Package sysinfo wraps the Linux sysinfo(2) syscall to provide system
// memory and swap statistics needed for ZSwap usage summary calculations.
//
// Key Components:
//   - SysInfo struct: Holds all fields from struct sysinfo
//   - Get(): Calls sysinfo(2) and returns a populated SysInfo
//   - Float64 accessors: Byte-scaled convenience methods for each field
//   - IsSwapAvailable(): Whether any swap space exists
//
// Dependencies:
//   - syscall: Linux sysinfo(2) wrapper
//   - os: Page size detection
//
// Error Types:
//   - Returned by Get() if sysinfo(2) fails
package sysinfo

import (
	"fmt"
	"os"
	"syscall"
)

// SysInfo holds the snapshot from sysinfo(2).
type SysInfo struct {
	Uptime       int64
	TotalRam     uint64
	FreeRam      uint64
	SharedRam    uint64
	BufferedRam  uint64
	TotalSwap    uint64
	FreeSwap     uint64
	Processes    uint16
	TotalHighMem uint64
	FreeHighMem  uint64
	MemUnit      uint32
	PageSize     int
}

/*
Get calls sysinfo(2) and returns a populated SysInfo.

    returns:
          *SysInfo: system memory and swap snapshot
          error:    non-nil if sysinfo(2) fails
*/
func Get() (*SysInfo, error) {
	var s syscall.Sysinfo_t
	if err := syscall.Sysinfo(&s); err != nil {
		return nil, fmt.Errorf("sysinfo: %w", err)
	}
	return &SysInfo{
		Uptime:       int64(s.Uptime),
		TotalRam:     uint64(s.Totalram),
		FreeRam:      uint64(s.Freeram),
		SharedRam:    uint64(s.Sharedram),
		BufferedRam:  uint64(s.Bufferram),
		TotalSwap:    uint64(s.Totalswap),
		FreeSwap:     uint64(s.Freeswap),
		Processes:    s.Procs,
		TotalHighMem: uint64(s.Totalhigh),
		FreeHighMem:  uint64(s.Freehigh),
		MemUnit:      uint32(s.Unit),
		PageSize:     os.Getpagesize(),
	}, nil
}

// TotalRamBytes returns total physical RAM in bytes.
func (s *SysInfo) TotalRamBytes() float64 {
	return float64(s.TotalRam) * float64(s.MemUnit)
}

// FreeRamBytes returns free physical RAM in bytes.
func (s *SysInfo) FreeRamBytes() float64 {
	return float64(s.FreeRam) * float64(s.MemUnit)
}

// SharedRamBytes returns shared memory in bytes.
func (s *SysInfo) SharedRamBytes() float64 {
	return float64(s.SharedRam) * float64(s.MemUnit)
}

// BufferedRamBytes returns buffered memory in bytes.
func (s *SysInfo) BufferedRamBytes() float64 {
	return float64(s.BufferedRam) * float64(s.MemUnit)
}

// TotalSwapBytes returns total swap space in bytes.
func (s *SysInfo) TotalSwapBytes() float64 {
	return float64(s.TotalSwap) * float64(s.MemUnit)
}

// FreeSwapBytes returns free swap space in bytes.
func (s *SysInfo) FreeSwapBytes() float64 {
	return float64(s.FreeSwap) * float64(s.MemUnit)
}

// TotalHighMemBytes returns total high memory in bytes.
func (s *SysInfo) TotalHighMemBytes() float64 {
	return float64(s.TotalHighMem) * float64(s.MemUnit)
}

// FreeHighMemBytes returns free high memory in bytes.
func (s *SysInfo) FreeHighMemBytes() float64 {
	return float64(s.FreeHighMem) * float64(s.MemUnit)
}

/*
IsSwapAvailable reports whether any swap space is configured.

    returns:
          bool: true if TotalSwap > 0
*/
func (s *SysInfo) IsSwapAvailable() bool {
	return s.TotalSwap != 0
}
