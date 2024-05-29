package main

import (
	"log"
	"os"

	"github.com/jukuly/ss_mach_mo/internal/cli/out"
	"github.com/jukuly/ss_mach_mo/internal/gui"
	"github.com/jukuly/ss_mach_mo/internal/model"
)

const GUI = true

func main() {
	if GUI {
		out.SetLogger(log.New(&gui.Console{}, "", 0))
	} else {
		out.SetLogger(log.New(os.Stdout, "", 0))
	}

	var sensors *[]model.Sensor = &[]model.Sensor{}
	var gateway *model.Gateway = &model.Gateway{}
	model.LoadSensors(model.SENSORS_FILE, sensors)
	err := model.LoadSettings(model.GATEWAY_FILE, gateway)
	if err != nil {
		out.Log("Error loading Gateway settings. Use 'config --id <gateway-id>' and 'config --password <gateway-password>' to set the Gateway settings.")
	}

	//err = server.Init(sensors, gateway)
	if err != nil {
		out.Error(err)
		//panic("Error initializing server")
	}

	//err = server.StartAdvertising()
	if err != nil {
		out.Error(err)
		//panic("Error starting advertising")
	}

	//in.Start(sensors, gateway)

	gui.Start()
}
