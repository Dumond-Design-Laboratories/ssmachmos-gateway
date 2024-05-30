package main

import (
	"github.com/jukuly/ss_mach_mo/server/internal/api"
	"github.com/jukuly/ss_mach_mo/server/internal/model"
	"github.com/jukuly/ss_mach_mo/server/internal/model/server"
	"github.com/jukuly/ss_mach_mo/server/internal/out"
)

func main() {
	var sensors *[]model.Sensor = &[]model.Sensor{}
	var gateway *model.Gateway = &model.Gateway{}
	model.LoadSensors(model.SENSORS_FILE, sensors)
	err := model.LoadSettings(gateway, model.GATEWAY_FILE)
	if err != nil {
		out.Log("Error loading Gateway settings. Run 'ssmachmos config --id <gateway-id>' and 'ssmachmos config --password <gateway-password>' to set the Gateway settings.")
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

	api.Start()
}
