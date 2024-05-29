package main

import (
	"github.com/jukuly/ss_mach_mo/internal/cli/in"
	"github.com/jukuly/ss_mach_mo/internal/cli/out"
	"github.com/jukuly/ss_mach_mo/internal/model"
	"github.com/jukuly/ss_mach_mo/internal/model/server"
)

func main() {
	var sensors *[]model.Sensor = &[]model.Sensor{}
	var gateway *model.Gateway = &model.Gateway{}
	model.LoadSensors(model.SENSORS_FILE, sensors)
	err := model.LoadSettings(model.GATEWAY_FILE, gateway)
	if err != nil {
		out.Log("Error loading Gateway settings. Use 'config --id <gateway-id>' and 'config --password <gateway-password>' to set the Gateway settings.")
	}

	err = server.Init(sensors, gateway)
	if err != nil {
		out.Error(err)
		//panic("Error initializing server")
	}

	err = server.StartAdvertising()
	if err != nil {
		out.Error(err)
		//panic("Error starting advertising")
	}

	in.Start(sensors, gateway)

}
