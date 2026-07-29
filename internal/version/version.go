// Package version provides product identity constants and kernel version
// detection for the zswapctl application.
//
// Key Components:
//   - Name, Version: Build-time product identity, set via ldflags
//   - KernelVersion(): Runtime kernel release string from uname(2)
//   - HeadersVersion(): Compile-time kernel headers version
//
// Dependencies:
//   - golang.org/x/sys/unix: uname(2) syscall for kernel release
//
// Error Types:
//   - Returned by KernelVersion() if uname(2) fails
package version

import (
	"fmt"
	"strings"

	"golang.org/x/sys/unix"
)

var (
	Name    = "zswapctl"
	Version = "1.0.0"
	Author  = "Dlcuy22"
)

// Kernel headers version constants, matching <linux/version.h> defines.
// Update these when building against newer kernel headers.
const (
	HeadersVersionMajor    = 6
	HeadersVersionPatchlevel = 15
	HeadersVersionSublevel   = 0
)

/*
KernelVersion returns the runtime kernel release string.

    returns:
          string: kernel release (e.g. "6.15.0-generic")
          error:  non-nil if uname(2) fails
*/
func KernelVersion() (string, error) {
	var u unix.Utsname
	if err := unix.Uname(&u); err != nil {
		return "", fmt.Errorf("uname: %w", err)
	}
	release := string(u.Release[:])
	release = strings.TrimRight(release, "\x00")
	release = strings.TrimRight(release, "-")
	return release, nil
}

/*
HeadersVersion returns the kernel headers version as a dotted string.

    returns:
          string: headers version (e.g. "6.15.0")
*/
func HeadersVersion() string {
	return fmt.Sprintf("%d.%d.%d", HeadersVersionMajor, HeadersVersionPatchlevel, HeadersVersionSublevel)
}
