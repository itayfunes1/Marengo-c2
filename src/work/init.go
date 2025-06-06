package work

func Init() {
	MasterListClientResp = make(chan string)
	MasterCreateBridgeResp = make(chan string)
	BridgeCameraResp = make(chan string)
	receiverFileJob = make(chan uint32, 1000)
	go receiveFileWorker()
}

func Stop() {
	close(MasterCreateBridgeResp)
	close(MasterListClientResp)
	close(BridgeCameraResp)
	close(receiverFileJob)
	if MasterConnection != nil {
		MasterConnection.Close()
	}
	if BridgeConnection != nil {
		BridgeConnection.Close()
	}
}
