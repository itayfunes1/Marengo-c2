package pty

import (
	"bytes"
	"sync"
)

type BufferWrap struct {
	Buf      bytes.Buffer
	MutexIn  sync.Mutex
	MutexOut sync.Mutex
}

func (b *BufferWrap) Write(p []byte) (int, error) {
	str := string(p)
	if str == lastCmd {
		return 0, nil
	}
	// fmt.Print(string(p))
	b.MutexIn.Lock()
	defer b.MutexIn.Unlock()
	n, err := b.Buf.Write([]byte(p))

	return n, err
}

func (b *BufferWrap) Read(p []byte) (int, error) {
	// b.Mutex.Lock()
	// defer b.Mutex.Unlock()
	b.MutexOut.Lock()
	defer b.MutexOut.Unlock()
	n, err := b.Buf.Read(p)
	// fmt.Println(2222, n, err)

	return n, err
}
