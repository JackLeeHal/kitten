//go:build darwin
// +build darwin

package os

func Fdatasync(fd uintptr) (err error) {
	return
}
