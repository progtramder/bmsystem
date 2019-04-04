// +build windows

package main

import (
	"fmt"
	"syscall"
)

func ColorPrint(s string, i int) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("SetConsoleTextAttribute")
	handle, _, _ := proc.Call(uintptr(syscall.Stdout), uintptr(i))
	fmt.Println(s)
	handle, _, _ = proc.Call(uintptr(syscall.Stdout), uintptr(7))
	CloseHandle := kernel32.NewProc("CloseHandle")
	CloseHandle.Call(handle)
}

func ColorRed(s string) {
	ColorPrint(s, 4|8)
}

func ColorGreen(s string) {
	ColorPrint(s, 2|8)
}

func ColorBlue(s string) {
	ColorPrint(s, 1|8)
}
