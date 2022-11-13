package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	setConsoleCursorPosition       = kernel32.NewProc("SetConsoleCursorPosition")
)

// progress marker
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
type timer struct {
	s      time.Time     // start time
	done   chan struct{} // for ticker end
	before int           // before display message size
}

// Start is returns timer and initialize.
func Start(msg *string) *timer {
	timer := &timer{
		s:    time.Now().Local(),
		done: make(chan struct{}),
	}

	fmt.Printf("[timer]start at %s\n\n", timer.s.Format("2006-01-02 15:04:05"))

	t := time.NewTicker(100 * time.Millisecond)
	go func() {
		defer t.Stop()
		cnt := 1
		for {
			select {
			case <-t.C:
				d := time.Now().Local().Sub(timer.s).Milliseconds()
				timer.print(fmt.Sprintf("%s please wait. %.1f sec", progress[cnt%len(progress)], float64(d)/1_000))

				cnt++
			case <-timer.done:
				timer.print("complete!")
				e := time.Now().Local()
				fmt.Fprintf(os.Stdout, "\n\n[timer]finish at %s\n", e.Format("2006-01-02 15:04:05"))
				fmt.Fprintf(os.Stdout, "[timer]latency is %d ms\n\n", e.Sub(timer.s).Milliseconds())
				return
			}

		}
	}()

	return timer
}

// print is returns the cursor to the beginning
// and displays message.
func (t *timer) print(msg string) {
	// add carriage return
	msg = "\r" + msg

	pos, _ := getCursorPos()
	pos.Y -= int16(len(msg))
	if pos.Y < 0 {
		pos.Y = 0
	}
	setCursorPos(pos)

	// Fill in the missing space with blanks
	// to remain shorter than the prev display.
	less := t.before - len(msg)
	if less < 0 {
		less = 0
	}
	fmt.Fprint(os.Stdout, msg+strings.Repeat(" ", less))
	t.before = len(msg)
}

// Finish is finished ticker and goroutine.
func (t *timer) Finish() {
	close(t.done)
}

func getCursorPos() (pos coordinates, err error) {
	var info consoleScreenBufferInfo
	_, _, e := syscall.SyscallN(procGetConsoleScreenBufferInfo.Addr(), 2, uintptr(syscall.Stdout), uintptr(unsafe.Pointer(&info)), 0)
	if e != 0 {
		return info.dwCursorPosition, error(e)
	}
	return info.dwCursorPosition, nil
}

func setCursorPos(pos coordinates) error {
	_, _, e := syscall.SyscallN(setConsoleCursorPosition.Addr(), 2, uintptr(syscall.Stdout), uintptr(uint32(uint16(pos.X))<<16|uint32(uint16(pos.Y))), 0)
	if e != 0 {
		return error(e)
	}
	return nil
}