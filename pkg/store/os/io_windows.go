//go:build windows
// +build windows

package os

const (
	O_NOATIME = 0 // darwin no O_NOATIME set to O_LARGEFILE
)
