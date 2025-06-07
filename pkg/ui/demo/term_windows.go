//go:build windows

package demo

import (
	"syscall"
	"unsafe"
)

var (
	kernel32DLL                = syscall.NewLazyDLL("kernel32.dll")
	getConsoleScreenBufferInfo = kernel32DLL.NewProc("GetConsoleScreenBufferInfo")
)

type (
	short int16
	word  uint16

	// Koordinaten im Konsolenfenster
	coord struct {
		x short
		y short
	}

	// Rechteck im Konsolenfenster
	smallRect struct {
		left   short
		top    short
		right  short
		bottom short
	}

	// Konsoleninfo-Struktur
	consoleScreenBufferInfo struct {
		size              coord
		cursorPosition    coord
		attributes        word
		window            smallRect
		maximumWindowSize coord
	}
)

// getTerminalSize implementiert die Terminalgrößenerkennung für Windows-Systeme
func getTerminalSize(fd uintptr) (width, height int, err error) {
	var info consoleScreenBufferInfo
	if ret, _, err := getConsoleScreenBufferInfo.Call(
		fd,
		uintptr(unsafe.Pointer(&info)),
	); ret == 0 {
		return 80, 24, err
	}

	// Berechne die Größe des Fensters
	width = int(info.window.right - info.window.left + 1)
	height = int(info.window.bottom - info.window.top + 1)
	return width, height, nil
}
