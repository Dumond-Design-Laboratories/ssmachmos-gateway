package in

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jukuly/ss_mach_mo/internal/model"
	"github.com/jukuly/ss_mach_mo/internal/view/out"
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

	options = options[:j]
	args = args[:k]

	switch tokens[0] {
	case "help":
		help(args)
	case "list":
		list(sensors)
	case "view":
		view(args, sensors)
	case "pair":
		pair(options, args, gateway)
	case "forget":
		forget(args, sensors)
	case "config":
		config(options, args, sensors, gateway)
	default:
		fmt.Printf("Unknown command: %s\n", tokens[0])
	}
}

func Start(sensors *[]model.Sensor, gateway *model.Gateway) {
	for {
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			out.Error(err)
		}
		handleInput(text, sensors, gateway)
	}
}
