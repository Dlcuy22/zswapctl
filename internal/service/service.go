// Package service manages systemd service installation and management for zswapctl.
//
// Key Components:
//   - Install(): Writes embedded config and systemd unit, reloads systemd, and enables/starts the service
//
// Dependencies:
//   - github.com/zswap-go/zswapctl/assets: Embedded configuration and systemd service unit
//   - github.com/zswap-go/zswapctl/internal/sysutil: Privilege check
//   - os, os/exec: File operations and systemctl execution
//
// Error Types:
//   - Returns error if file creation, permission check, or systemctl command fails
package service

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/zswap-go/zswapctl/assets"
	"github.com/zswap-go/zswapctl/internal/sysutil"
)

const (
	// ConfigDir is the target directory for zswapctl configuration file
	ConfigDir = "/etc/zswapctl"
	// ConfigFile is the target path for zswapctl configuration file
	ConfigFile = "/etc/zswapctl/zswapctl.conf"
	// ServiceFile is the target path for zswapctl systemd service unit
	ServiceFile = "/etc/systemd/system/zswapctl.service"
)

/*
Install installs built-in configuration and systemd service unit, then enables and starts the service.

    returns:
          error: non-nil if privilege check, file writing, or systemctl execution fails
*/
func Install() error {
	if !sysutil.IsRoot() {
		return fmt.Errorf("the requested action requires super-user privileges")
	}

	fmt.Println("Installing zswapctl configuration and systemd service...")

	// Create config directory
	if err := os.MkdirAll(ConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", ConfigDir, err)
	}

	// Write zswapctl.conf
	if err := os.WriteFile(ConfigFile, assets.ConfigFS, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", ConfigFile, err)
	}
	fmt.Printf("Created configuration file: %s\n", ConfigFile)

	// Write zswapctl.service
	if err := os.WriteFile(ServiceFile, assets.ServiceFS, 0644); err != nil {
		return fmt.Errorf("failed to write service file %s: %w", ServiceFile, err)
	}
	fmt.Printf("Created systemd service file: %s\n", ServiceFile)

	// Reload systemd daemon
	fmt.Println("Reloading systemd manager configuration...")
	cmdReload := exec.Command("systemctl", "daemon-reload")
	if out, err := cmdReload.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to execute systemctl daemon-reload: %w (%s)", err, string(out))
	}

	// Enable and start zswapctl service
	fmt.Println("Enabling and starting zswapctl service...")
	cmdEnable := exec.Command("systemctl", "enable", "--now", "zswapctl.service")
	if out, err := cmdEnable.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to enable and start zswapctl service: %w (%s)", err, string(out))
	}

	fmt.Println("Successfully installed, enabled, and started zswapctl service.")
	return nil
}
