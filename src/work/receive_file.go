package work

import (
	"log"
	"time"
)

func receiveFileWorker() {
	for fileId := range receiverFileJob {
		handleReceiveFile(fileId)
	}
}

func handleReceiveFile(fileId uint32) {
	value, ok := FileMap.Load(fileId)
	if !ok {
		return
	}
	fileContent := value.(*(FileContent))
	timeOut := make(chan bool)
	incTimeOut := make(chan bool)
	isTimeout := false
	isDone := false
	go func() {
		start := time.Now()
		timeOutDuration := 10 * time.Second
		for {
			select {
			case <-time.After(1 * time.Second):
				now := time.Now()
				if now.Sub(start) > timeOutDuration {
					timeOut <- true
					return
				}
				if isDone {
					return
				}

			case <-incTimeOut:
				start = time.Now()

			}
		}
	}()
	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				if isDone {
					return
				}
				log.Printf("\r\nProgress: [%d]/[%d]=[%.2f]", fileContent.Size, fileContent.TotalSize, float64(fileContent.Size)/float64(fileContent.TotalSize)*100)

			}
		}
	}()
	for {
		select {
		case <-timeOut:
			isTimeout = true
		case content := <-fileContent.DataChan:
			incTimeOut <- true
			fileContent.Mutex.Lock()
			fileContent.Data[content.Sequence] = content.Data
			fileContent.Size = fileContent.Size + uint32(len(content.Data))
			fileContent.Mutex.Unlock()

			// fmt.Println(content.Sequence, fileContent.Size, fileContent.TotalSize)
		}
		if fileContent.Size == fileContent.TotalSize && fileContent.TotalSize != 0 {
			log.Println("\nReceived full file, save file", fileContent.Name, fileContent.Size)
			isDone = true
			fileContent.Done <- true
			break
		}
		if isTimeout {
			isDone = true
			log.Println("Timeout")
			fileContent.Done <- false
			break
		}
	}
	// select {
	// case <-fileContent.Done:
	// 	fmt.Println("Content fully downloaded")
	// }
}
