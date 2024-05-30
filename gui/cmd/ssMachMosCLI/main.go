package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jukuly/ss_mach_mo/gui/internal/commands"
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

	conn, err := commands.OpenConnection()
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer conn.Close()

	switch as[0] {
	case "help":
		commands.Help(args, conn)
	case "list":
		commands.List(conn)
	case "view":
		commands.View(args, conn)
	case "pair":
		commands.Pair(args, conn)
	case "forget":
		commands.Forget(args, conn)
	case "config":
		commands.Config(options, args, conn)
	case "stop":
		commands.Stop(conn)
	default:
		fmt.Printf("Unknown command: %s\n", as[0])
	}
}
