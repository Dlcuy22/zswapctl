// Package sysutil provides low-level system utility functions for privilege
// checking and page size detection.
//
// Key Components:
//   - IsRoot(): Checks if the current process has superuser privileges
//   - GetPageSize(): Returns the system memory page size in bytes
//
// Dependencies:
//   - os: uid inspection and page size
//
// Error Types:
//   None; all functions are infallible on Linux
package sysutil

import "os"

/*
IsRoot checks whether the current process is running as superuser.

    returns:
          bool: true if uid is 0
*/
func IsRoot() bool {
	return os.Getuid() == 0
}

/*
GetPageSize returns the system memory page size in bytes.

    returns:
          int: page size in bytes (typically 4096)
*/
func GetPageSize() int {
	return os.Getpagesize()
}
