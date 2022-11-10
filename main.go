package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	setConsoleCursorPosition = kernel32.NewProc("SetConsoleCursorPosition")
)

var progress = []string{"-", "/", "|", "\\"}

type smallRect struct {
	Left, Top, Right, Bottom int16
}
type coordinates struct {
	X, Y int16
}
type consoleScreenBufferInfo struct {
	dwSize              coordinates
	dwCursorPosition    coordinates
	wAttributes         int16
	srWindow            smallRect
	dwMaximumWindowSize coordinates
}

func main() {
	t := time.NewTicker(100*time.Millisecond)
	go func () {
		cnt := 1
		for {
			<-t.C

			cnt++
			pos, _ := getCursorPos()
			pos.Y -= 13
			resetCursorPos(pos)
			fmt.Printf("\r%s please wait.", progress[cnt%len(progress)])
		}
	}()

	fmt.Scanln()
}

func getCursorPos() (pos coordinates, err error) {
	var info consoleScreenBufferInfo
	_, _, e := syscall.SyscallN(procGetConsoleScreenBufferInfo.Addr(), 2, uintptr(syscall.Stdout), uintptr(unsafe.Pointer(&info)), 0)
	if e != 0 {
		return info.dwCursorPosition, error(e)
	}
	return info.dwCursorPosition, nil
}

func resetCursorPos(pos coordinates) error {
	_, _, e := syscall.SyscallN(setConsoleCursorPosition.Addr(), 2, uintptr(syscall.Stdout), uintptr(uint32(uint16(pos.X))<<16|uint32(uint16(pos.Y))), 0)
	if e != 0 {
		return error(e)
	}
	return nil
}
