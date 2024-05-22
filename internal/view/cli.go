package view

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jukuly/ss_mach_mo/internal/model"
)

func handleInput(input string, sensors *[]model.Sensor, gateway *model.Gateway) {
	input = strings.TrimSpace(input)
	tokens := strings.Split(input, " ")
	if len(tokens) == 0 || tokens[0] == "" {
		return
	}
	options := make([]string, len(tokens)-1)
	args := make([]string, len(tokens)-1)

	var (
		j int
		k int
	)
	for i, token := range tokens {
		if i == 0 {
			continue
		}

		if strings.HasPrefix(token, "--") {
			options[j] = token
			j++
		} else {
			args[k] = token
			k++
		}
	}

	switch tokens[0] {
	case "help":
		help(args)
	case "list":
		list(sensors)
	case "view":
		view(args, sensors)
	case "pair":
		pair()
	case "forget":
		forget(args, sensors)
	case "config":
		config(options, args, sensors, gateway)
	default:
		fmt.Printf("Unknown command: %s\n", tokens[0])
	}
}

func Error(err error) {
	Log(err.Error())
}

func Log(msg string) {
	fmt.Printf("[%s] %s\n", time.Now().Format(time.RFC3339), msg)
}

func Start(sensors *[]model.Sensor, gateway *model.Gateway) {
	for {
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			Error(err)
		}
		handleInput(text, sensors, gateway)
	}
}
