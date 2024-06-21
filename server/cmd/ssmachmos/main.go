package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jukuly/ss_machmos/server/internal/api"
	"github.com/jukuly/ss_machmos/server/internal/cli"
	"github.com/jukuly/ss_machmos/server/internal/model"
	"github.com/jukuly/ss_machmos/server/internal/model/server"
	"github.com/jukuly/ss_machmos/server/internal/out"
)

func serve() {
	// if the server is already running, stop it
	out.Logger.Println("Closing running instances...")
	conn, err := cli.OpenConnection()
	if err == nil {
		conn.Write([]byte("STOP\x00"))
		conn.Close()
	}

	out.Logger.Println("Loading local config...")
	var sensors *[]model.Sensor = &[]model.Sensor{}
	var gateway *model.Gateway = &model.Gateway{
		HTTPEndpoint: "https://openphm.org/gateway_data",
	}
	model.LoadSensors(model.SENSORS_FILE, sensors)
	err = model.LoadSettings(gateway, model.GATEWAY_FILE)
	if err != nil {
		out.Logger.Println("Error loading Gateway settings. Run 'ssmachmos config --id <gateway-id>' and 'ssmachmos config --password <gateway-password>' to set the Gateway settings.")
	}

	out.Logger.Println("Starting bluetooth advertisement...")
	err = server.Init(sensors, gateway)
	if err != nil {
		out.Logger.Println("Error:", err)
	} else {
		err = server.StartAdvertising()
		if err != nil {
			out.Logger.Println("Error:", err)
		}
	}

	if err == nil {
		out.Logger.Println("Done initializing server.")
	} else {
		out.Logger.Println("Done initializing server with errors.")
	}
	api.Start()
}

func main() {

	// the user must provide at least one argument (the command)
	as := os.Args[1:]
	if len(as) == 0 {
		fmt.Println("Usage: ssmachmos <command> [options] [arguments]")
		return
	}

	options := make([]string, len(as)-1)
	args := make([]string, len(as)-1)

	var (
		i int
		j int
	)
	// skip the first argument (the command)
	for _, a := range as[1:] {
		if strings.HasPrefix(a, "--") {
			options[i] = a
			i++
		} else {
			args[j] = a
			j++
		}
	}

	options = options[:i]
	args = args[:j]

	if as[0] == "serve" {
		if len(options) > 0 && options[0] == "--no-console" {
			process, err := os.StartProcess(os.Args[0], []string{os.Args[0], "serve"}, &os.ProcAttr{
				Files: []*os.File{nil, nil, nil},
				Env:   os.Environ(),
			})
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			err = process.Release()
			if err != nil {
				fmt.Println("Error:", err)
			}
			fmt.Println("Server started.")
			fmt.Println("To stop the server, run 'ssmachmos stop'.")
			fmt.Println("To view the live server logs, run 'ssmachmos logs'.")
		} else {
			serve()
		}
		return
	}

	if as[0] == "help" {
		cli.Help(args)
		return
	}

	// open a unix domain socket connection to the server
	conn, err := cli.OpenConnection()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	go cli.Listen(conn)
	defer conn.Close()

	switch as[0] {
	case "logs":
		cli.Logs(conn)
	case "list":
		cli.List(conn)
	case "view":
		cli.View(options, args, conn)
	case "pair":
		cli.Pair(args, conn)
	case "forget":
		cli.Forget(args, conn)
	case "config":
		cli.Config(options, args, conn)
	case "stop":
		cli.Stop(conn)
	default:
		fmt.Printf("Unknown command: %s\n", as[0])
	}
}
