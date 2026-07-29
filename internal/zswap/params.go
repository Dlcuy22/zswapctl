// Package zswap provides read/write access to ZSwap kernel module parameters
// via sysfs and debug statistics via debugfs.
//
// Key Components:
//   - SetValue/GetValue: Generic parameter accessors with validation
//   - SetEnabled, GetEnabled, etc.: Typed wrappers for each parameter
//   - IsAvailable(): Checks if the zswap sysfs directory exists
//   - Debug stat accessors: GetPoolTotalSize, GetStoredPages, etc.
//   - IsDebugAvailable(): Checks if the zswap debugfs directory exists
//
// Dependencies:
//   - os: File I/O for sysfs/debugfs nodes
//   - regexp: Validation patterns for boolean and range values
//   - strconv: uint64 parsing for debug statistics
//
// Error Types:
//   - Returned by SetValue on validation failure or write/verify errors
//   - Returned by debug getters on parse errors
package zswap

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const ModuleParametersPath = "/sys/module/zswap/parameters/"

// ModuleDebugPath is the debugfs mount point for zswap statistics.
const ModuleDebugPath = "/sys/kernel/debug/zswap/"

var (
	boolRe  = regexp.MustCompile(`^[YN]$`)
	rangeRe = regexp.MustCompile(`^\d{1,2}$|^100$`)

	paramDefs = []paramDef{
		{name: "enabled", kind: paramBool},
		{name: "same_filled_pages_enabled", kind: paramBool},
		{name: "max_pool_percent", kind: paramRange},
		{name: "compressor", kind: paramNonEmpty},
		{name: "zpool", kind: paramNonEmpty},
		{name: "accept_threshold_percent", kind: paramRange},
		{name: "non_same_filled_pages_enabled", kind: paramBool},
		{name: "exclusive_loads", kind: paramBool},
		{name: "shrinker_enabled", kind: paramBool},
	}
)

type paramKind int

const (
	paramBool     paramKind = iota // Y or N
	paramRange                     // integer 0-100
	paramNonEmpty                  // any non-empty string
)

type paramDef struct {
	name string
	kind paramKind
}

/*
IsAvailable checks whether the zswap sysfs parameters directory exists.

    returns:
          bool: true if /sys/module/zswap/parameters/ is a directory
*/
func IsAvailable() bool {
	info, err := os.Stat(ModuleParametersPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

/*
IsDebugAvailable checks whether the zswap debugfs directory exists.

    returns:
          bool: true if /sys/kernel/debug/zswap/ is a directory
*/
func IsDebugAvailable() bool {
	info, err := os.Stat(ModuleDebugPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func readValue(name string) (string, error) {
	data, err := os.ReadFile(ModuleParametersPath + name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func writeValue(name, value string) error {
	return os.WriteFile(ModuleParametersPath+name, []byte(value), 0644)
}

func validateBool(name, value string) error {
	if !boolRe.MatchString(value) {
		return fmt.Errorf("the requested value for the option %q is incorrect (only Y or N are supported)", name)
	}
	return nil
}

func validateRange(name, value string) error {
	if !rangeRe.MatchString(value) {
		return fmt.Errorf("the requested value for the option %q is out of range [0..100]", name)
	}
	return nil
}

func validateNonEmpty(name, value string) error {
	if value == "" {
		return fmt.Errorf("the requested value for the option %q is empty", name)
	}
	return nil
}

/*
SetValue validates and writes a ZSwap module parameter.

    params:
          name:  parameter file name under /sys/module/zswap/parameters/
          value: new value to write

    returns:
          error: non-nil on validation failure, write error, or verification mismatch

    note: After writing, the value is read back to confirm the kernel accepted it.
*/
func SetValue(name, value string) error {
	for _, pd := range paramDefs {
		if pd.name == name {
			switch pd.kind {
			case paramBool:
				if err := validateBool(name, value); err != nil {
					return err
				}
			case paramRange:
				if err := validateRange(name, value); err != nil {
					return err
				}
			case paramNonEmpty:
				if err := validateNonEmpty(name, value); err != nil {
					return err
				}
			}
			break
		}
	}

	fullPath := ModuleParametersPath + name
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("configuring the option %q is not possible on the current kernel", name)
	}

	oldValue, _ := readValue(name)

	if err := writeValue(name, value); err != nil {
		return fmt.Errorf("failed to write option %q: %w", name, err)
	}

	newValue, err := readValue(name)
	if err != nil {
		return fmt.Errorf("failed to verify option %q: %w", name, err)
	}
	if newValue != value {
		return fmt.Errorf("failed to set the option %q to %q, current value %q remains unchanged", name, value, oldValue)
	}

	fmt.Printf("The option %q has been set to a new value of %q (old value was %q).\n", name, value, oldValue)
	return nil
}

/*
GetValue reads a ZSwap module parameter value.

    params:
          name: parameter file name under /sys/module/zswap/parameters/

    returns:
          string: current value, empty string if parameter does not exist
          error:  non-nil on read failure
*/
func GetValue(name string) (string, error) {
	fullPath := ModuleParametersPath + name
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", nil
	}
	return readValue(name)
}

// Typed setters
func SetEnabled(v string) error                    { return SetValue("enabled", v) }
func SetSameFilledPagesEnabled(v string) error     { return SetValue("same_filled_pages_enabled", v) }
func SetMaxPoolPercent(v string) error             { return SetValue("max_pool_percent", v) }
func SetCompressor(v string) error                 { return SetValue("compressor", v) }
func SetZpool(v string) error                      { return SetValue("zpool", v) }
func SetAcceptThresholdPercent(v string) error      { return SetValue("accept_threshold_percent", v) }
func SetNonSameFilledPagesEnabled(v string) error  { return SetValue("non_same_filled_pages_enabled", v) }
func SetExclusiveLoads(v string) error             { return SetValue("exclusive_loads", v) }
func SetShrinkerEnabled(v string) error            { return SetValue("shrinker_enabled", v) }

// Typed getters
func GetEnabled() (string, error)                    { return GetValue("enabled") }
func GetSameFilledPagesEnabled() (string, error)     { return GetValue("same_filled_pages_enabled") }
func GetMaxPoolPercent() (string, error)             { return GetValue("max_pool_percent") }
func GetCompressor() (string, error)                 { return GetValue("compressor") }
func GetZpool() (string, error)                      { return GetValue("zpool") }
func GetAcceptThresholdPercent() (string, error)      { return GetValue("accept_threshold_percent") }
func GetNonSameFilledPagesEnabled() (string, error)  { return GetValue("non_same_filled_pages_enabled") }
func GetExclusiveLoads() (string, error)             { return GetValue("exclusive_loads") }
func GetShrinkerEnabled() (string, error)            { return GetValue("shrinker_enabled") }

// readDebugValue reads a single debugfs value as uint64.
func readDebugValue(name string) (uint64, error) {
	data, err := os.ReadFile(ModuleDebugPath + name)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %q: %w", name, err)
	}
	return v, nil
}

// getDebugUint64 reads a debugfs value, returning 0 if the file does not exist.
func getDebugUint64(name string) (uint64, error) {
	fullPath := ModuleDebugPath + name
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return 0, nil
	}
	return readDebugValue(name)
}

// Debug stat accessors
func GetPoolLimitHit() (uint64, error)          { return getDebugUint64("pool_limit_hit") }
func GetPoolTotalSize() (uint64, error)         { return getDebugUint64("pool_total_size") }
func GetRejectAllocFail() (uint64, error)       { return getDebugUint64("reject_alloc_fail") }
func GetRejectCompressPoor() (uint64, error)    { return getDebugUint64("reject_compress_poor") }
func GetRejectKmemCacheFail() (uint64, error)   { return getDebugUint64("reject_kmemcache_fail") }
func GetRejectReclaimFail() (uint64, error)     { return getDebugUint64("reject_reclaim_fail") }
func GetRejectCompressFail() (uint64, error)    { return getDebugUint64("reject_compress_fail") }
func GetDecompressFail() (uint64, error)        { return getDebugUint64("decompress_fail") }
func GetSameFilledPages() (uint64, error)       { return getDebugUint64("same_filled_pages") }
func GetStoredPages() (uint64, error)           { return getDebugUint64("stored_pages") }
func GetWrittenBackPages() (uint64, error)      { return getDebugUint64("written_back_pages") }
func GetIncompressiblePages() (uint64, error)   { return getDebugUint64("stored_incompressible_pages") }
