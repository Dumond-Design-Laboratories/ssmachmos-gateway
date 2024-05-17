package view

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func handleInput(input string, updated chan<- bool) {
	input = strings.TrimSpace(input)
	tokens := strings.Split(input, " ")
	if len(tokens) == 0 || tokens[0] == "" {
		return
	}
	options := make([]string, len(tokens)-1)
	args := make([]string, len(tokens)-1)

	for i, token := range tokens {
		if i == 0 {
			continue
		}

		if strings.HasPrefix(token, "--") {
			options[i-1] = token
		} else {
			args[i-1] = token
		}
	}

	switch tokens[0] {
	case "help":
		help(args)
	case "list":
		list()
	case "view":
		view(args)
	case "pair":
		pair()
	case "forget":
		forget(args)
	case "config":
		config(options, args, updated)
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

func Start(updated chan<- bool) {
	for {
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			Error(err)
		}
		handleInput(text, updated)
	}
}
