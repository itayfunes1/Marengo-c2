package server

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func getInputInt() (int, error) {
	var choice int
	reader := bufio.NewReader(os.Stdin)
	command, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}

	choice, err = strconv.Atoi(strings.TrimSpace(command))
	if err != nil {
		return 0, err
	}
	return choice, nil
}

func getInputString() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	command, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(command), nil
}
