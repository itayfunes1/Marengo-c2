package work

import (
	"fmt"
	"log"
	"server-tcp-go/src/handler/message"
	"server-tcp-go/src/utils"
	"strings"
)

func ListClient() {
	clientList := getListClient()
	if clientList == nil && len(clientList) == 0 {
		log.Println("No client")
	}

	for i, client := range clientList {
		fmt.Printf("Client %d - Ip %s\n", i+1, client)
	}
}

func getListClient() []string {
	mess := message.NewMsg(
		message.MessageTypeGetListClient,
		[]byte(""),
		utils.GetLocalIP(),
	)

	messByte, err := mess.Encode()
	if err != nil {
		log.Println(err)
		return nil
	}

	MasterConnection.Write(messByte)

	resp := <-MasterListClientResp
	if resp == "" {
		return nil
	}
	clientListResp := strings.Split(resp, "\n")
	var clientList []string
	for _, c := range clientListResp {
		if c == "" {
			continue
		}
		clientList = append(clientList, c)
	}
	return clientList
}

func ListClientResp(resp string) {
	MasterListClientResp <- resp
}
