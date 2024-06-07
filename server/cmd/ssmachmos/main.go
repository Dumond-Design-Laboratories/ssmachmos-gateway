package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jukuly/ss_mach_mo/server/internal/api"
	"github.com/jukuly/ss_mach_mo/server/internal/cli"
	"github.com/jukuly/ss_mach_mo/server/internal/model"
	"github.com/jukuly/ss_mach_mo/server/internal/model/server"
	"github.com/jukuly/ss_mach_mo/server/internal/out"
)

func serve() {
	conn, err := cli.OpenConnection()
	if err == nil {
		conn.Write([]byte("STOP\n"))
		conn.Close()
	}

	var sensors *[]model.Sensor = &[]model.Sensor{}
	var gateway *model.Gateway = &model.Gateway{}
	model.LoadSensors(model.SENSORS_FILE, sensors)
	err = model.LoadSettings(gateway, model.GATEWAY_FILE)
	if err != nil {
		out.Logger.Print("Error loading Gateway settings. Run 'ssmachmos config --id <gateway-id>' and 'ssmachmos config --password <gateway-password>' to set the Gateway settings.")
	}

	err = server.Init(sensors, gateway)
	if err != nil {
		out.Logger.Print(err)
	}
	if err == nil {
		err = server.StartAdvertising()
		if err != nil {
			out.Logger.Print(err)
		}
	}

	api.Start()
}

func main() {
	as := os.Args[1:]
	if len(as) == 0 {
		fmt.Println("Usage: ssmachmos <command> [options] [arguments]")
		return
	}

	if as[0] == "serve" {
		serve()
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

	conn, err := cli.OpenConnection()
	go cli.Listen(conn)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer conn.Close()

	switch as[0] {
	case "help":
		cli.Help(args, conn)
	case "list":
		cli.List(conn)
	case "view":
		cli.View(args, conn)
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
