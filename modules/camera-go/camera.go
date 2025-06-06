package camerago

import (
	"encoding/binary"
	"fmt"
	"marengo/modules/screenshot"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"
)

// https://www.magnumdb.com/
const (
	WM_CAP_DRIVER_CONNECT = 1034
	WM_CAP_DISCONNECT     = 1035
	//WM_CAP_COPY           = 1054
	//WM_CAP_GRAB_FRAME         = 1084
	//WM_CAP_FILE_SAVEDIBA      = 1049
	//WM_CAP_FILE_SAVEAS        = 1047
	WM_CAP_SINGLE_FRAME_OPEN  = 1094
	WM_CAP_SINGLE_FRAME       = 1096
	WM_CAP_SET_CAPTURE_FILE   = 1044
	WM_CAP_SINGLE_FRAME_CLOSE = 1095
	WM_CAP_SET_CALLBACK_FRAME = 1029
	WM_CAP_GET_VIDEOFORMAT    = 1068
)

var (
	user32            = syscall.MustLoadDLL("user32.dll")
	SendMessageProc   = user32.MustFindProc("SendMessageA")
	PostMessageProc   = user32.MustFindProc("PostMessageA")
	DestroyWindowProc = user32.MustFindProc("DestroyWindow")
	//EnumClipboardFormatsProc = user32.MustFindProc("EnumClipboardFormats")
	//OpenClipboardProc        = user32.MustFindProc("OpenClipboard")
	//CloseClipboardProc       = user32.MustFindProc("CloseClipboard")
)

var (
	avicap32                    = syscall.MustLoadDLL("avicap32.dll")
	capCreateCaptureWindowAProc = avicap32.MustFindProc("capCreateCaptureWindowA")
	//capGetDriverDescriptionAProc = avicap32.MustFindProc("capGetDriverDescriptionA")
)

type VIDEOHDR struct {
	lpData         uintptr
	dwBufferLength uint32
	dwBytesUsed    uint32
	dwTimeCaptured uint32
	dwUser         uintptr
	dwFlags        uint32
	dwReserved     [4]uintptr
}

type BITMAPFILEHEADER struct {
	bfType      [2]byte
	bfSize      uint32
	bfReserved1 uint16
	bfReserved2 uint16
	bfOffBits   uint32
}

type BITMAPINFOHEADER struct {
	biSize          uint32
	biWidth         int32
	biHeight        int32
	biPlanes        uint16
	biBitCount      uint16
	biCompression   uint32
	biSizeImage     uint32
	biXPelsPerMeter int32
	biYPelsPerMeter int32
	biClrUsed       uint32
	biClrImportant  uint32
}

type BITMAPINFO struct {
	bmiHeader BITMAPINFOHEADER
	bmiColors [1]RGBQUAD
}

type RGBQUAD struct {
	rgbBlue     byte
	rgbGreen    byte
	rgbRed      byte
	rgbReserved byte
}

func capCreateCaptureWindowA(lpszWindowName string, dwStyle uint32, x, y, nWidth, nHeight int32, hWndParent, nID uintptr) uintptr {
	file := &append([]byte(lpszWindowName), 0)[0]
	ret, _, _ := capCreateCaptureWindowAProc.Call(
		uintptr(unsafe.Pointer(file)),
		uintptr(dwStyle),
		uintptr(x),
		uintptr(y),
		uintptr(nWidth),
		uintptr(nHeight),
		hWndParent,
		nID,
	)
	return ret
}

/*func capGetDriverDescriptionA(wDriverIndex int, lpszName *byte, cbName int, lpszVer *byte, cbVer int) uintptr {
	ret, _, _ := capGetDriverDescriptionAProc.Call(
		uintptr(wDriverIndex),
		uintptr(unsafe.Pointer(lpszName)),
		uintptr(cbName),
		uintptr(unsafe.Pointer(lpszVer)),
		uintptr(cbVer),
	)
	return ret
}*/

func SendMessage(hwnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	ret1, ret2, err := SendMessageProc.Call(hwnd, uintptr(msg), wParam, lParam)
	fmt.Printf("%d, ret1 %d, ret2 %d\n", msg, ret1, ret2)
	fmt.Printf("%s\n", err)
	return ret1
}

func PostMessage(hwnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	ret1, ret2, err := PostMessageProc.Call(hwnd, uintptr(msg), wParam, lParam)
	fmt.Printf("%d, ret1 %d, ret2 %d\n", msg, ret1, ret2)
	fmt.Printf("%s\n", err)
	return ret1
}

/*func EnumClipboardFormats(format uint32) uintptr {
	ret, _, _ := EnumClipboardFormatsProc.Call(uintptr(format))
	return ret
}

func OpenClipboard() uintptr {
	r, _, _ := OpenClipboardProc.Call(0)
	return r
}

func CloseClipboard() uintptr {
	r, _, _ := CloseClipboardProc.Call()
	return r
}*/

/*func PowerShellMethod() {
	out, err := exec.Command("PowerShell", "-Command", "Add-Type", "-AssemblyName", "System.Windows.Forms;$clip=[Windows.Forms.Clipboard]::GetImage();if ($clip -ne $null) { $clip.Save('image.bmp') };").CombinedOutput()
	fmt.Println(out)
	fmt.Println(err)
	if err != nil {
		fmt.Println("fuck")
		panic(out)
	}
}*/

func CaptureCallback(hwnd uintptr, lpvhdr *VIDEOHDR) uint64 {
	//println("start of callback")
	data := unsafe.Slice((*byte)(unsafe.Pointer(lpvhdr.lpData)), lpvhdr.dwBytesUsed)
	var videoFormat BITMAPINFO
	size := SendMessage(hwnd, WM_CAP_GET_VIDEOFORMAT, unsafe.Sizeof(videoFormat), 0)
	SendMessage(hwnd, WM_CAP_GET_VIDEOFORMAT, uintptr(size), uintptr(unsafe.Pointer(&videoFormat)))

	tempDir := os.Getenv("TEMP")
	/* We save it as a .png even though it can be a jpeg or a bitmap, because most
	photo viewers open it anyway */
	savePath := filepath.Join(tempDir, "frame.png")
	file, err := os.Create(savePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	/* check if it's compressed. if it is then just write the compressed data
	because it already has the header. if not, add the bitmap headers and write */
	if videoFormat.bmiHeader.biCompression != 0 {
		binary.Write(file, binary.LittleEndian, data)
		return 0
	}
	var header BITMAPFILEHEADER
	header.bfType[0] = 'B'
	header.bfType[1] = 'M'
	header.bfSize = uint32(14 + 40 + len(data))
	header.bfOffBits = 54
	//fmt.Printf("BMH bfType: %s\nBMH bfSize: %d\nBMH bfOffBits: %d\n", header.bfType, header.bfSize, header.bfOffBits)
	//fmt.Printf("BMI biSize: %d\nBMI biWidth: %d\nBMI biHeight: %d\nBMI biPlanes: %d\nBMI biBitCount: %d\nBMI biCompression: %d\nBMI biSizeImage: %d\nBMI biXPelsPerMeter: %d\nBMI biYPelsPerMeter: %d\nBMI biClrUsed: %d\nBMI biClrImportant: %d\n", videoFormat.bmiHeader.biSize, videoFormat.bmiHeader.biWidth, videoFormat.bmiHeader.biHeight, videoFormat.bmiHeader.biPlanes, videoFormat.bmiHeader.biBitCount, videoFormat.bmiHeader.biCompression, videoFormat.bmiHeader.biSizeImage, videoFormat.bmiHeader.biXPelsPerMeter, videoFormat.bmiHeader.biYPelsPerMeter, videoFormat.bmiHeader.biClrUsed, videoFormat.bmiHeader.biClrImportant)
	binary.Write(file, binary.LittleEndian, &header)
	binary.Write(file, binary.LittleEndian, &videoFormat.bmiHeader)
	binary.Write(file, binary.LittleEndian, data)
	return 0
}

func capture() {
	/*szDeviceName := make([]byte, 80)
	szDeviceVersion := make([]byte, 80)
	capGetDriverDescriptionA(0, &szDeviceName[0], 80, &szDeviceVersion[0], 80)
	fmt.Println(string(szDeviceName))
	fmt.Println(string(szDeviceVersion))*/
	var hCaptureWnd uintptr = capCreateCaptureWindowA("xxxxx", 0, 0, 0, 350, 350, 0, 0)
	SendMessage(hCaptureWnd, WM_CAP_DRIVER_CONNECT, 0, 0)
	time.Sleep(2000 * time.Millisecond)
	//fmt.Println("slept 2s")
	SendMessage(hCaptureWnd, WM_CAP_SET_CALLBACK_FRAME, 0, syscall.NewCallback(CaptureCallback))
	name := "vid"
	fileName := &append([]byte(name), 0)[0]
	SendMessage(hCaptureWnd, WM_CAP_SET_CAPTURE_FILE, 0, uintptr(unsafe.Pointer(fileName)))
	SendMessage(hCaptureWnd, WM_CAP_SINGLE_FRAME_OPEN, 0, 0)
	SendMessage(hCaptureWnd, WM_CAP_SINGLE_FRAME, 0, 0)
	SendMessage(hCaptureWnd, WM_CAP_SINGLE_FRAME_CLOSE, 0, 0)
	//save := SendMessage(hCaptureWnd, WM_CAP_FILE_SAVEAS, 0, uintptr(unsafe.Pointer(fileName)))
	SendMessage(hCaptureWnd, WM_CAP_DISCONNECT, 0, 0)

	DestroyWindowProc.Call(hCaptureWnd)

	os.Remove("vid")

	/*if save == 0 {
		fmt.Println("Failed to save image with WM_CAP_FILE_SAVEDIBA. Falling back to the Go library method.")
		err := clipboard.Init()
		if err != nil {
			fmt.Println("Failed to save image with the Go library method. Falling back to the PowerShell method.")
			PowerShellMethod()
		}

		file, err := os.Create("image.bmp")
		if err != nil {
			fmt.Println("Failed to save image with the Go library method. Falling back to the PowerShell method.")
			file.Close()
			PowerShellMethod()
		}

		clip := clipboard.Read(clipboard.FmtImage)

		if clip != nil {
			file.Write(clip)
		} else {
			fmt.Println("Failed to save image with the Go library method. Falling back to the PowerShell method.")
			file.Close()
			PowerShellMethod()
		}

		file.Close()
	}*/

	/*runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	OpenClipboard()
	defer CloseClipboard()

	for i := 0; i <= 10; i++ {
		fmt.Printf("Clipboard format: %d\n", EnumClipboardFormats(uint32(i)))
	}*/
}

func PostCapture() {
	/*szDeviceName := make([]byte, 80)
	szDeviceVersion := make([]byte, 80)
	capGetDriverDescriptionA(0, &szDeviceName[0], 80, &szDeviceVersion[0], 80)
	fmt.Println(string(szDeviceName))
	fmt.Println(string(szDeviceVersion))*/
	window := uintptr(screenshot.GetDesktopWindow())
	var hCaptureWnd uintptr = capCreateCaptureWindowA("cutie", 0, 0, 0, 350, 350, window, 0)
	SendMessage(hCaptureWnd, WM_CAP_DRIVER_CONNECT, 0, 0)
	time.Sleep(2000 * time.Millisecond)
	//fmt.Println("slept 2s")
	PostMessage(hCaptureWnd, WM_CAP_SET_CALLBACK_FRAME, 0, syscall.NewCallback(CaptureCallback))
	name := "vid"
	fileName := &append([]byte(name), 0)[0]
	PostMessage(hCaptureWnd, WM_CAP_SET_CAPTURE_FILE, 0, uintptr(unsafe.Pointer(fileName)))
	PostMessage(hCaptureWnd, WM_CAP_SINGLE_FRAME_OPEN, 0, 0)
	PostMessage(hCaptureWnd, WM_CAP_SINGLE_FRAME, 0, 0)
	PostMessage(hCaptureWnd, WM_CAP_SINGLE_FRAME_CLOSE, 0, 0)
	//save := SendMessage(hCaptureWnd, WM_CAP_FILE_SAVEAS, 0, uintptr(unsafe.Pointer(fileName)))
	SendMessage(hCaptureWnd, WM_CAP_DISCONNECT, 0, 0)

	DestroyWindowProc.Call(hCaptureWnd)

	os.Remove("vid")

}
