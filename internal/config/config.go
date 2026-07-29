// Package config loads ZSwap configuration from INI-style config files.
//
// Key Components:
//   - Config struct: Holds all configurable ZSwap parameters
//   - Load(): Parses an INI file and returns a populated Config
//
// Dependencies:
//   - gopkg.in/ini.v1: INI file parser
//
// Error Types:
//   - Returned by Load() if the file is missing or malformed
package config

import (
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

// Config holds ZSwap parameters parsed from an INI config file.
type Config struct {
	Enabled                    string
	SameFilledPagesEnabled     string
	MaxPoolPercent             string
	Compressor                 string
	Zpool                      string
	AcceptThresholdPercent     string
	NonSameFilledPagesEnabled  string
	ExclusiveLoads             string
	ShrinkerEnabled            string
}

/*
Load reads and parses a ZSwap INI configuration file.

    params:
          path: absolute path to the .conf file

    returns:
          *Config: populated configuration struct
          error:   non-nil if file is missing or malformed
*/
func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("the specified configuration file does not exist: %s", path)
	}

	f, err := ini.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	sec := f.Section("zswap")
	if sec == nil {
		return nil, fmt.Errorf("missing [zswap] section in config file")
	}

	cfg := &Config{}
	if sec.HasKey("enabled") {
		cfg.Enabled = sec.Key("enabled").String()
	}
	if sec.HasKey("same_filled_pages_enabled") {
		cfg.SameFilledPagesEnabled = sec.Key("same_filled_pages_enabled").String()
	}
	if sec.HasKey("max_pool_percent") {
		cfg.MaxPoolPercent = sec.Key("max_pool_percent").String()
	}
	if sec.HasKey("compressor") {
		cfg.Compressor = sec.Key("compressor").String()
	}
	if sec.HasKey("zpool") {
		cfg.Zpool = sec.Key("zpool").String()
	}
	if sec.HasKey("accept_threshold_percent") {
		cfg.AcceptThresholdPercent = sec.Key("accept_threshold_percent").String()
	}
	if sec.HasKey("non_same_filled_pages_enabled") {
		cfg.NonSameFilledPagesEnabled = sec.Key("non_same_filled_pages_enabled").String()
	}
	if sec.HasKey("exclusive_loads") {
		cfg.ExclusiveLoads = sec.Key("exclusive_loads").String()
	}
	if sec.HasKey("shrinker_enabled") {
		cfg.ShrinkerEnabled = sec.Key("shrinker_enabled").String()
	}

	return cfg, nil
}
