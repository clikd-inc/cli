//go:build darwin || linux

package demo

import (
	"syscall"
	"unsafe"
)

// getTerminalSize implementiert die Terminalgrößenerkennung für Unix-Systeme
func getTerminalSize(fd uintptr) (width, height int, err error) {
	var dimensions [4]uint16
	if _, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		fd,
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&dimensions)),
		0, 0, 0,
	); errno != 0 {
		return 80, 24, errno
	}
	return int(dimensions[1]), int(dimensions[0]), nil
}
