// Package main implements the zswapctl command-line interface.
//
// Key Components:
//   - Cobra root command with flag parsing
//   - Three input modes: CLI flags, config file, environment variables
//   - Four stats display modes: combined, settings, summary, debug
//
// Dependencies:
//   - github.com/spf13/cobra: CLI framework
//   - github.com/zswap-go/zswapctl/internal/config: INI config loader
//   - github.com/zswap-go/zswapctl/internal/sysinfo: sysinfo(2) wrapper
//   - github.com/zswap-go/zswapctl/internal/sysutil: privilege and page size
//   - github.com/zswap-go/zswapctl/internal/version: product identity
//   - github.com/zswap-go/zswapctl/internal/zswap: sysfs/debugfs access
//
// Error Types:
//   - All errors propagated via cobra RunE, printed and exit(1) in main
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zswap-go/zswapctl/internal/config"
	"github.com/zswap-go/zswapctl/internal/service"
	"github.com/zswap-go/zswapctl/internal/sysinfo"
	"github.com/zswap-go/zswapctl/internal/sysutil"
	"github.com/zswap-go/zswapctl/internal/version"
	"github.com/zswap-go/zswapctl/internal/zswap"
)

var (
	flagConfig  string
	flagEnv     bool
	flagStats   int
	flagVerbose bool
	flagInstall bool

	flagEnabled                    string
	flagSameFilledPagesEnabled     string
	flagMaxPoolPercent             string
	flagCompressor                 string
	flagZpool                      string
	flagAcceptThresholdPercent     string
	flagNonSameFilledPagesEnabled  string
	flagExclusiveLoads             string
	flagShrinkerEnabled            string
)

func main() {
	rootCmd := &cobra.Command{
		Use:           "zswapctl",
		Short:         "Command-line tool to control the ZSwap kernel module options",
		Long:          "Zswapctl is a command-line tool to control the zswap kernel module options on the fly.\nRewritten to Go by Dlcuy22.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          runRoot,
	}

	rootCmd.Version = version.Version
	rootCmd.SetVersionTemplate(fmt.Sprintf("%s version %s\nRewritten to Go by %s\n", version.Name, version.Version, version.Author))

	rootCmd.Flags().StringVar(&flagConfig, "config", "", "get options from the configuration file instead of the cmdline")
	rootCmd.Flags().BoolVar(&flagEnv, "env", false, "get options from the environment variables instead of the cmdline")
	rootCmd.Flags().IntVar(&flagStats, "stats", -1, "show statistics and current settings of ZSwap kernel module (0=all, 1=settings, 2=summary, 3=debug)")
	rootCmd.Flags().BoolVar(&flagVerbose, "verbose", false, "enable verbose mode to display additional information")
	rootCmd.Flags().BoolVar(&flagInstall, "install", false, "install built-in config and systemd service, then enable and start zswapctl")

	rootCmd.Flags().StringVarP(&flagEnabled, "enabled", "e", "", "enable or disable the ZSwap kernel module")
	rootCmd.Flags().StringVarP(&flagSameFilledPagesEnabled, "same_filled_pages_enabled", "s", "", "enable or disable memory pages deduplication")
	rootCmd.Flags().StringVarP(&flagMaxPoolPercent, "max_pool_percent", "p", "", "the maximum percentage of memory that the compressed pool can occupy")
	rootCmd.Flags().StringVarP(&flagCompressor, "compressor", "c", "", "the algorithm used to compress memory pages")
	rootCmd.Flags().StringVarP(&flagZpool, "zpool", "z", "", "the kernel's zpool type")
	rootCmd.Flags().StringVarP(&flagAcceptThresholdPercent, "accept_threshold_percent", "a", "", "the threshold at which ZSwap would start accepting pages again after it became full")
	rootCmd.Flags().StringVarP(&flagNonSameFilledPagesEnabled, "non_same_filled_pages_enabled", "n", "", "enable or disable accepting non same filled memory pages")
	rootCmd.Flags().StringVarP(&flagExclusiveLoads, "exclusive_loads", "x", "", "enable or disable entries invalidation when memory pages are loaded from compressed pool")
	rootCmd.Flags().StringVarP(&flagShrinkerEnabled, "shrinker_enabled", "r", "", "enable or disable pool shrinking based on memory pressure")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

/*
runRoot is the cobra RunE handler that dispatches to the appropriate mode.

    params:
          cmd:  the cobra Command instance
          args: positional arguments (unused)

    returns:
          error: non-nil on any failure
*/
func runRoot(cmd *cobra.Command, args []string) error {
	hasCmdLineFlags := cmd.Flags().Changed("enabled") ||
		cmd.Flags().Changed("same_filled_pages_enabled") ||
		cmd.Flags().Changed("max_pool_percent") ||
		cmd.Flags().Changed("compressor") ||
		cmd.Flags().Changed("zpool") ||
		cmd.Flags().Changed("accept_threshold_percent") ||
		cmd.Flags().Changed("non_same_filled_pages_enabled") ||
		cmd.Flags().Changed("exclusive_loads") ||
		cmd.Flags().Changed("shrinker_enabled")

	if !hasCmdLineFlags && !flagEnv && flagConfig == "" && flagStats < 0 && !flagInstall {
		return cmd.Help()
	}

	if flagInstall {
		return service.Install()
	}

	if flagStats >= 0 {
		return printStats(flagStats)
	}

	if !sysutil.IsRoot() {
		return fmt.Errorf("the requested action requires super-user privileges")
	}

	if flagConfig != "" {
		return executeConfig(flagConfig)
	}

	if flagEnv {
		return executeEnv()
	}

	return executeCmdLine()
}

/*
executeConfig loads settings from an INI file and applies them.

    params:
          path: absolute path to the config file

    returns:
          error: first validation or write error encountered
*/
func executeConfig(path string) error {
	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

	var firstErr error
	set := func(name, value string, setter func(string) error) {
		if value == "" {
			return
		}
		if err := setter(value); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	set("enabled", cfg.Enabled, zswap.SetEnabled)
	set("same_filled_pages_enabled", cfg.SameFilledPagesEnabled, zswap.SetSameFilledPagesEnabled)
	set("max_pool_percent", cfg.MaxPoolPercent, zswap.SetMaxPoolPercent)
	set("compressor", cfg.Compressor, zswap.SetCompressor)
	set("zpool", cfg.Zpool, zswap.SetZpool)
	set("accept_threshold_percent", cfg.AcceptThresholdPercent, zswap.SetAcceptThresholdPercent)
	set("non_same_filled_pages_enabled", cfg.NonSameFilledPagesEnabled, zswap.SetNonSameFilledPagesEnabled)
	set("exclusive_loads", cfg.ExclusiveLoads, zswap.SetExclusiveLoads)
	set("shrinker_enabled", cfg.ShrinkerEnabled, zswap.SetShrinkerEnabled)

	return firstErr
}

/*
executeEnv reads ZSwap settings from environment variables.

    returns:
          error: first validation or write error encountered
*/
func executeEnv() error {
	envMap := map[string]func(string) error{
		"ZSWAP_ENABLED_VALUE":                      zswap.SetEnabled,
		"ZSWAP_SAME_FILLED_PAGES_ENABLED_VALUE":    zswap.SetSameFilledPagesEnabled,
		"ZSWAP_MAX_POOL_PERCENT_VALUE":              zswap.SetMaxPoolPercent,
		"ZSWAP_COMPRESSOR_VALUE":                    zswap.SetCompressor,
		"ZSWAP_ZPOOL_VALUE":                         zswap.SetZpool,
		"ZSWAP_ACCEPT_THRESHOLD_PERCENT_VALUE":      zswap.SetAcceptThresholdPercent,
		"ZSWAP_NON_SAME_FILLED_PAGES_ENABLED_VALUE": zswap.SetNonSameFilledPagesEnabled,
		"ZSWAP_EXCLUSIVE_LOADS_VALUE":               zswap.SetExclusiveLoads,
		"ZSWAP_SHRINKER_ENABLED_VALUE":              zswap.SetShrinkerEnabled,
	}

	var firstErr error
	for key, setter := range envMap {
		val := os.Getenv(key)
		if val == "" {
			continue
		}
		if err := setter(val); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

/*
executeCmdLine applies ZSwap settings from CLI flags.

    returns:
          error: first validation or write error encountered
*/
func executeCmdLine() error {
	type flagPair struct {
		value  string
		setter func(string) error
	}

	pairs := []flagPair{
		{flagEnabled, zswap.SetEnabled},
		{flagSameFilledPagesEnabled, zswap.SetSameFilledPagesEnabled},
		{flagMaxPoolPercent, zswap.SetMaxPoolPercent},
		{flagCompressor, zswap.SetCompressor},
		{flagZpool, zswap.SetZpool},
		{flagAcceptThresholdPercent, zswap.SetAcceptThresholdPercent},
		{flagNonSameFilledPagesEnabled, zswap.SetNonSameFilledPagesEnabled},
		{flagExclusiveLoads, zswap.SetExclusiveLoads},
		{flagShrinkerEnabled, zswap.SetShrinkerEnabled},
	}

	var firstErr error
	for _, p := range pairs {
		if p.value == "" {
			continue
		}
		if err := p.setter(p.value); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

/*
printStats dispatches to the correct display mode.

    params:
          mode: 0=combined, 1=settings, 2=summary, 3=debug

    returns:
          error: non-nil on invalid mode or display failure
*/
func printStats(mode int) error {
	switch mode {
	case 0:
		return printCombined()
	case 1:
		return printSettings()
	case 2:
		return printSummary()
	case 3:
		return printDebugInfo()
	default:
		return fmt.Errorf("incorrect value of the --stats command-line option was specified")
	}
}

func checkModuleLoaded() error {
	if !zswap.IsAvailable() {
		return fmt.Errorf("ZSwap kernel module is not loaded or access to sysfs is denied")
	}
	return nil
}

func checkDebugAvailable() error {
	if !zswap.IsDebugAvailable() {
		return fmt.Errorf("ZSwap is not running or access to debugfs is denied")
	}
	return nil
}

/*
printSettings displays current ZSwap module parameter values.
Prints "NOT SUPPORTED" for missing parameters when verbose mode is active.
*/
func printSettings() error {
	if err := checkModuleLoaded(); err != nil {
		return err
	}

	type setting struct {
		name string
		get  func() (string, error)
	}

	settings := []setting{
		{"ZSwap enabled", zswap.GetEnabled},
		{"Same filled pages enabled", zswap.GetSameFilledPagesEnabled},
		{"Maximum pool percentage", zswap.GetMaxPoolPercent},
		{"Compression algorithm", zswap.GetCompressor},
		{"Kernel's zpool type", zswap.GetZpool},
		{"Accept threshold percentage", zswap.GetAcceptThresholdPercent},
		{"Non same filled pages enabled", zswap.GetNonSameFilledPagesEnabled},
		{"Exclusive loads", zswap.GetExclusiveLoads},
		{"Shrinker enabled", zswap.GetShrinkerEnabled},
	}

	for _, s := range settings {
		val, err := s.get()
		if err != nil {
			return err
		}
		if val != "" {
			fmt.Printf("%s: %s\n", s.name, val)
		} else if flagVerbose {
			fmt.Printf("%s: NOT SUPPORTED\n", s.name)
		}
	}
	return nil
}

/*
printDebugInfo displays ZSwap debugfs statistics.
Requires debugfs access; returns an error if unavailable.
*/
func printDebugInfo() error {
	if err := checkDebugAvailable(); err != nil {
		return err
	}

	type debugStat struct {
		name string
		get  func() (uint64, error)
	}

	stats := []debugStat{
		{"Pool limit hit", zswap.GetPoolLimitHit},
		{"Pool total size", zswap.GetPoolTotalSize},
		{"Reject allocation failures", zswap.GetRejectAllocFail},
		{"Reject compression poor", zswap.GetRejectCompressPoor},
		{"Reject Kmemcache failures", zswap.GetRejectKmemCacheFail},
		{"Reject reclaim failures", zswap.GetRejectReclaimFail},
		{"Reject compression failures", zswap.GetRejectCompressFail},
		{"Decompression failures", zswap.GetDecompressFail},
		{"Same filled pages count", zswap.GetSameFilledPages},
		{"Stored pages count", zswap.GetStoredPages},
		{"Written back pages count", zswap.GetWrittenBackPages},
		{"Incompressible pages count", zswap.GetIncompressiblePages},
	}

	for _, s := range stats {
		val, err := s.get()
		if err != nil {
			return err
		}
		if val > 0 {
			fmt.Printf("%s: %d\n", s.name, val)
		} else if flagVerbose {
			fmt.Printf("%s: NOT SUPPORTED\n", s.name)
		}
	}
	return nil
}

/*
printSummary displays ZSwap usage summary with pool size, stored data,
and compression ratio.
*/
func printSummary() error {
	if err := checkDebugAvailable(); err != nil {
		return err
	}

	si, err := sysinfo.Get()
	if err != nil {
		return err
	}

	if !si.IsSwapAvailable() {
		return fmt.Errorf("ZSwap is not functional due to missing swap file or partition")
	}

	poolSize, _ := zswap.GetPoolTotalSize()
	if poolSize == 0 {
		return fmt.Errorf("ZSwap is not working. The pool is empty")
	}

	storedPages, _ := zswap.GetStoredPages()

	storedSize := float64(storedPages) * float64(si.PageSize)
	poolSizeMB := float64(poolSize) / 1048576.0
	memTotalPercent := float64(poolSize) / si.TotalRamBytes() * 100.0
	storedSizeMB := storedSize / 1048576.0
	swapUsedBytes := si.TotalSwapBytes() - si.FreeSwapBytes()
	var swapUsedPercent float64
	if swapUsedBytes > 0 {
		swapUsedPercent = storedSize / swapUsedBytes * 100.0
	}
	var compressionRatio float64
	if poolSizeMB > 0 {
		compressionRatio = storedSizeMB / poolSizeMB
	}

	fmt.Printf("Pool: %.2f MiB (%.1f%% of MemTotal)\n", poolSizeMB, memTotalPercent)
	fmt.Printf("Stored: %.2f MiB (%.1f%% of SwapUsed)\n", storedSizeMB, swapUsedPercent)
	fmt.Printf("Compression ratio: %.2f\n", compressionRatio)
	return nil
}

/*
printCombined prints all three sections: settings, summary, debug.
Errors in individual sections are printed but do not abort the overall output.
*/
func printCombined() error {
	sections := []struct {
		title string
		fn    func() error
	}{
		{"ZSWAP KERNEL MODULE SETTINGS:", printSettings},
		{"ZSWAP KERNEL MODULE USAGE SUMMARY:", printSummary},
		{"ZSWAP KERNEL MODULE DEBUG INFO:", printDebugInfo},
	}

	for _, s := range sections {
		fmt.Println(s.title)
		if err := s.fn(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		fmt.Println()
	}
	return nil
}
