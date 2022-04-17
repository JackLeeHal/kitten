//go:build windows
// +build windows

package os

func Fdatasync(fd uintptr) (err error) {
	return
}
