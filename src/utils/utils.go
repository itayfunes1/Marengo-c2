package utils

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func GetEnv(envName, defaultValue string) string {
	res := os.Getenv(envName)
	if res == "" {
		return defaultValue
	}
	return res
}

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "unknown"
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "unknown"
	}

	return strings.TrimSpace(string(body))
}

// // GetLocalIP returns the non loopback local IP of the host
// func GetLocalIP() string {
// 	addrs, err := net.InterfaceAddrs()
// 	if err != nil {
// 		return ""
// 	}
// 	for _, address := range addrs {
// 		// log.Println(address)
// 		// check the address type and if it is not a loopback the display it
// 		ipnet, ok := address.(*net.IPNet)
// 		// if ok && !ipnet.IP.IsLoopback() {
// 		if ok {
// 			if ipnet.IP.To4() != nil {
// 				return ipnet.IP.String()
// 			}
// 		}
// 	}
// 	return ""
// }
