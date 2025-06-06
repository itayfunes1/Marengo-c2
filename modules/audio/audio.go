package audio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/moutend/go-wav"
	"github.com/moutend/go-wca/pkg/wca"
)

func CaptureAudio(filename string, dura int) (err error) {
	if dura > 3600 {
		//fmt.Println("Request recording cannot over 1h. Aborting...")
		return
	}
	duration := time.Duration(dura) * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	audio, err := captureSharedEventDriven(ctx, duration)
	if err != nil {
		return
	}
	file, err := wav.Marshal(audio)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(filename, file, 0644)
	if err != nil {
		return
	}
	//fmt.Println("Successfully capture with filename", filename)
	return
}

func captureSharedEventDriven(ctx context.Context, duration time.Duration) (audio *wav.File, err error) {
	if err = ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return
	}
	defer ole.CoUninitialize()

	var mmde *wca.IMMDeviceEnumerator
	if err = wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &mmde); err != nil {
		return
	}
	defer mmde.Release()

	var mmd *wca.IMMDevice
	if err = mmde.GetDefaultAudioEndpoint(wca.ECapture, wca.EConsole, &mmd); err != nil {
		return
	}
	defer mmd.Release()

	var ps *wca.IPropertyStore
	if err = mmd.OpenPropertyStore(wca.STGM_READ, &ps); err != nil {
		return
	}
	defer ps.Release()

	var pv wca.PROPVARIANT
	if err = ps.GetValue(&wca.PKEY_Device_FriendlyName, &pv); err != nil {
		return
	}
	//fmt.Printf("Capturing audio from: %s\n", pv.String())
	var pv1 wca.PROPVARIANT
	if err = ps.GetValue(&wca.PKEY_DeviceInterface_FriendlyName, &pv1); err != nil {
		return
	}
	var ac *wca.IAudioClient
	if err = mmd.Activate(wca.IID_IAudioClient, wca.CLSCTX_ALL, nil, &ac); err != nil {
		return
	}
	defer ac.Release()

	var wfx *wca.WAVEFORMATEX
	if err = ac.GetMixFormat(&wfx); err != nil {
		return
	}
	defer ole.CoTaskMemFree(uintptr(unsafe.Pointer(wfx)))

	wfx.WFormatTag = 1
	wfx.NBlockAlign = (wfx.WBitsPerSample / 8) * wfx.NChannels
	wfx.NAvgBytesPerSec = wfx.NSamplesPerSec * uint32(wfx.NBlockAlign)
	wfx.CbSize = 0

	if audio, err = wav.New(int(wfx.NSamplesPerSec), int(wfx.WBitsPerSample), int(wfx.NChannels)); err != nil {
		return
	}

	//fmt.Println("--------")
	//fmt.Printf("Format: PCM %d bit signed integer\n", wfx.WBitsPerSample)
	//fmt.Printf("Rate: %d Hz\n", wfx.NSamplesPerSec)
	//fmt.Printf("Channels: %d\n", wfx.NChannels)
	//fmt.Println("--------")

	var defaultPeriod wca.REFERENCE_TIME
	var minimumPeriod wca.REFERENCE_TIME
	//var latency time.Duration
	if err = ac.GetDevicePeriod(&defaultPeriod, &minimumPeriod); err != nil {
		return
	}
	//latency = time.Duration(int(defaultPeriod) * 100)

	//fmt.Println("Default period: ", defaultPeriod)
	//fmt.Println("Minimum period: ", minimumPeriod)
	//fmt.Println(latency)

	if err = ac.Initialize(wca.AUDCLNT_SHAREMODE_SHARED, wca.AUDCLNT_STREAMFLAGS_EVENTCALLBACK, defaultPeriod, 0, wfx, nil); err != nil {
		//fmt.Println("ac.Initialize error: ", err)
		return
	}

	audioReadyEvent := wca.CreateEventExA(0, 0, 0, wca.EVENT_MODIFY_STATE|wca.SYNCHRONIZE)
	defer wca.CloseHandle(audioReadyEvent)

	if err = ac.SetEventHandle(audioReadyEvent); err != nil {
		//fmt.Println("SetEventHandle error: ", err)
		return
	}

	var bufferFrameSize uint32
	if err = ac.GetBufferSize(&bufferFrameSize); err != nil {
		//fmt.Println("GetBufferSize error: ", err)
		return
	}
	//fmt.Printf("Allocated buffer size: %d\n", bufferFrameSize)

	var acc *wca.IAudioCaptureClient
	if err = ac.GetService(wca.IID_IAudioCaptureClient, &acc); err != nil {
		//fmt.Println("GetService error: ", err)
		return
	}
	defer acc.Release()

	if err = ac.Start(); err != nil {
		//fmt.Println("ac.Start(): ", err)
		return
	}
	//fmt.Println("Start capturing with shared event driven mode")
	if duration <= 0 {
		//fmt.Println("Press Ctrl-C to stop capturing")
	}

	var output = []byte{}
	var offset int
	var isCapturing bool = true
	var currentDuration time.Duration
	var b *byte
	var data *byte
	var availableFrameSize uint32
	var flags uint32
	var devicePosition uint64
	var qcpPosition uint64
	//var padding uint32

	errorChan := make(chan error, 1)

	for {
		if !isCapturing {
			close(errorChan)
			break
		}
		go func() {
			errorChan <- watchEvent(ctx, audioReadyEvent)
		}()

		select {
		case <-ctx.Done():
			isCapturing = false
			<-errorChan
			break
		case err = <-errorChan:
			currentDuration = time.Duration(float64(offset) / float64(wfx.WBitsPerSample/8) / float64(wfx.NChannels) / float64(wfx.NSamplesPerSec) * float64(time.Second))
			if duration != 0 && currentDuration > duration {
				isCapturing = false
				break
			}
			if err != nil {
				isCapturing = false
				break
			}
			if err = acc.GetBuffer(&data, &availableFrameSize, &flags, &devicePosition, &qcpPosition); err != nil {
				continue
			}
			if availableFrameSize == 0 {
				continue
			}

			start := unsafe.Pointer(data)
			lim := int(availableFrameSize) * int(wfx.NBlockAlign)
			buf := make([]byte, lim)

			for n := 0; n < lim; n++ {
				b = (*byte)(unsafe.Pointer(uintptr(start) + uintptr(n)))
				buf[n] = *b
			}

			offset += lim
			output = append(output, buf...)

			if err = acc.ReleaseBuffer(availableFrameSize); err != nil {
				return
			}
		}
	}

	io.Copy(audio, bytes.NewBuffer(output))
	//fmt.Println("Stop capturing")

	if err = ac.Stop(); err != nil {
		return
	}
	return
}

func watchEvent(ctx context.Context, event uintptr) (err error) {
	errorChan := make(chan error, 1)
	go func() {
		errorChan <- eventEmitter(event)
	}()
	select {
	case err = <-errorChan:
		close(errorChan)
		return
	case <-ctx.Done():
		err = ctx.Err()
		return
	}

}

func eventEmitter(event uintptr) (err error) {
	//if err = ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
	//	return
	//}
	dw := wca.WaitForSingleObject(event, wca.INFINITE)
	if dw != 0 {
		return fmt.Errorf("failed to watch event")
	}
	//ole.CoUninitialize()
	return
}
