// Package assets provides embedded default configuration and systemd service files.
//
// Key Components:
//   - ConfigFS: Default zswapctl.conf file content
//   - ServiceFS: Default zswapctl.service unit file content
//
// Dependencies:
//   - embed: Go standard library embed package
//
// Error Types:
//   None
package assets

import _ "embed"

// ConfigFS contains the embedded default configuration file.
//go:embed zswapctl.conf
var ConfigFS []byte

// ServiceFS contains the embedded default systemd service unit file.
//go:embed zswapctl.service
var ServiceFS []byte
