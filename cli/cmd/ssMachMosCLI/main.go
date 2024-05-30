package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jukuly/ss_mach_mo/cli/internal/commands"
)

func main() {
	as := os.Args[1:]
	if len(as) == 0 {
		fmt.Println("Usage: ssmachmos <command> [options] [arguments]")
		return
	}

	options := make([]string, len(as)-1)
	args := make([]string, len(as)-1)

	var (
		j int
		k int
	)
	for i, a := range as {
		if i == 0 {
			continue
		}

		if strings.HasPrefix(a, "--") {
			options[j] = a
			j++
		} else {
			args[k] = a
			k++
		}
	}

	options = options[:j]
	args = args[:k]

	switch as[0] {
	case "help":
		commands.Help(args)
	case "list":
		commands.List()
	case "view":
		commands.View(args)
	case "pair":
		commands.Pair(args)
	case "forget":
		commands.Forget(args)
	case "config":
		commands.Config(options, args)
	case "stop":
		commands.Stop()
	default:
		fmt.Printf("Unknown command: %s\n", as[0])
	}
}
