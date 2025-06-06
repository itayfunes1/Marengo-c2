package pty

// Windows pty example
// https://devblogs.microsoft.com/commandline/windows-command-line-introducing-the-windows-pseudo-console-conpty/

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var kernel32Path = filepath.Join(os.Getenv("windir"), "System32", "kernel32.dll")

var (
	hKernel32            = syscall.NewLazyDLL(kernel32Path)
	fCreatePseudoConsole = hKernel32.NewProc("CreatePseudoConsole")
	fResizePseudoConsole = hKernel32.NewProc("ResizePseudoConsole")
	fClosePseudoConsole  = hKernel32.NewProc("ClosePseudoConsole")
)

type PTYModel struct {
	BufIn      BufferWrap
	BufOut     BufferWrap
	ResultChan chan string
}

const (
	S_OK uintptr = 0
)

type HandleIO struct {
	handle syscall.Handle
}

type COORD struct {
	X, Y int16
}

func (c *COORD) Pack() uintptr {
	return uintptr((int32(c.Y) << 16) | int32(c.X))
}

type HPCON syscall.Handle

func (h *HandleIO) Read(p []byte) (int, error) {
	// fmt.Println("1111Read", string(p[:]), "22222")
	var numRead uint32 = 0
	err := syscall.ReadFile(h.handle, p, &numRead, nil)
	return int(numRead), err
}

func (h *HandleIO) Write(p []byte) (int, error) {
	// fmt.Println("1111Write", string(p[:]))
	var numWritten uint32 = 0
	err := syscall.WriteFile(h.handle, p, &numWritten, nil)
	return int(numWritten), err
}

func (h *HandleIO) Close() error {
	return syscall.CloseHandle(h.handle)
}

func ClosePseudoConsole(hPc HPCON) {
	fClosePseudoConsole.Call(uintptr(hPc))
}

func ResizePseudoConsole(hPc HPCON, coord *COORD) error {
	ret, _, _ := fResizePseudoConsole.Call(uintptr(hPc), coord.Pack())
	if ret != S_OK {
		return fmt.Errorf("ResizePseudoConsole failed with status 0x%x", ret)
	}
	return nil
}

func CreatePseudoConsole(hIn, hOut syscall.Handle) (HPCON, error) {
	var hPc HPCON
	coord := &COORD{80, 40}
	ret, _, _ := fCreatePseudoConsole.Call(
		coord.Pack(),
		uintptr(hIn),
		uintptr(hOut),
		0,
		uintptr(unsafe.Pointer(&hPc)))
	if ret != S_OK {
		return 0, fmt.Errorf("CreatePseudoConsole() failed with status 0x%x", ret)
	}
	return hPc, nil
}

var lastCmd = ""

func (pty *PTYModel) SendCmd(cmd string) {
	// fmt.Println(cmd)
	pty.BufIn.Write([]byte(cmd + "\n"))
	lastCmd = cmd
}

func (pty *PTYModel) GetResult() {
	src := &pty.BufOut
	var buf []byte
	size := 1024
	buf = make([]byte, size)
	resultTotal := ""
	for {
		var str string
		var err error
		time.Sleep(200 * time.Millisecond)
		for {
			time.Sleep(200 * time.Millisecond)
			nw, er := src.Read(buf)
			str = str + string(buf[0:nw])
			if er != nil {
				if er != errors.New("EOF") {
					err = er
				}
				break
			}
		}
		if len(strings.TrimSpace(str)) == 0 {
			continue
		}
		if err != nil && err.Error() != "EOF" {
			fmt.Println(err)
			continue
		}
		// fmt.Println(str)
		resultTotal = resultTotal + str
		if strings.TrimSpace(str) != "" {
			arr := getBlockResult(&resultTotal)
			go func() {
				for _, value := range arr {
					pty.ResultChan <- value
				}

			}()
		}

	}

}

var pattern = `^([A-Za-z]:[^>]+>)\s*`

func validateBlock(str string) bool {
	return !strings.Contains(str, "Microsoft Corporation. All rights reserved.")
}

func getBlockResult(str *string) []string {
	regex := regexp.MustCompile(pattern)
	arr := []string{}
	scanner := bufio.NewScanner(strings.NewReader(*str))
	// block := ""
	var block = ""
	for scanner.Scan() {
		line := scanner.Text()

		match := regex.MatchString(line)

		if match {

			if block != "" && validateBlock(block) {
				arr = append(arr, block)
			}
			block = line + "\n"
			//
		} else {

			block = block + line + "\n"
		}
	}

	*str = block
	return arr
}

func (pty *PTYModel) Initalize() {

	var cmdIn, cmdOut syscall.Handle
	var ptyIn, ptyOut syscall.Handle
	if err := syscall.CreatePipe(&ptyIn, &cmdIn, nil, 0); err != nil {
		fmt.Printf("CreatePipe: %v", err)
	}
	if err := syscall.CreatePipe(&cmdOut, &ptyOut, nil, 0); err != nil {
		fmt.Printf("CreatePipe: %v", err)
	}

	cmd := exec.Command("cmd.exe")
	// Set the CREATE_NO_WINDOW flag to hide the console window
	const CREATE_NO_WINDOW = 0x08000000
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: CREATE_NO_WINDOW,
	}
	cmd.Stdin = &HandleIO{ptyIn}
	cmd.Stdout = &HandleIO{ptyOut}
	cmd.Stderr = &HandleIO{ptyOut}

	hPc, err := CreatePseudoConsole(ptyIn, ptyOut)
	if err != nil {
		log.Fatalf("CreatePseudoConsole %s", err)
	}
	// defer ClosePseudoConsole(hPc)
	ClosePseudoConsole(hPc)
	go io.Copy(&(pty.BufOut), &HandleIO{cmdOut})
	go func() {
		for {
			time.Sleep(200 * time.Millisecond)
			io.Copy(&HandleIO{cmdIn}, &(pty.BufIn))
		}

	}()
	go pty.GetResult()
	cmd.Run()
}
