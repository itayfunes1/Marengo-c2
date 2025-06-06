package server

import (
	"fmt"
	"log"
	"server-tcp-go/src/work"
	"strings"
)

const KeyPath = "./"

func PrintMenu() {
	for {
		fmt.Println(strings.Repeat("=", 40))
		fmt.Println("🧠  Marengo Command & Control Panel")
		fmt.Println(strings.Repeat("=", 40))
		fmt.Println("1️⃣  List connected agents")
		fmt.Println("2️⃣  Interact with an agent")
		// fmt.Println("3️⃣  Send command to agent using key")
		// fmt.Println("4️⃣  Save agent key to file")
		fmt.Println("0️⃣  Exit Marengo")
		fmt.Println(strings.Repeat("-", 40))
		fmt.Print("🟢  Enter your choice: ")

		var choice int
		choice, err := getInputInt()
		if err != nil {
			if err.Error() == "strconv.Atoi: parsing \"\": invalid syntax" {
				continue
			}
			log.Println("❌ Invalid choice. Please enter a number.", err)
			continue
		}

		fmt.Println()

		switch choice {
		case 0:
			log.Println("👋 Exiting Marengo...")
			return
		case 1:
			work.ListClient()
		case 2:
			for {
				exitSession := work.SendCommandToClient()
				if exitSession {
					fmt.Println("🔙 Returning to main menu...")
					break
				}
			}
		case 3:
			work.SendCommandToClientWithKey()
		case 4:
			work.SaveKeytoFile()
		default:
			log.Println("❌ Invalid choice. Please select a valid option.")
		}
	}
}
