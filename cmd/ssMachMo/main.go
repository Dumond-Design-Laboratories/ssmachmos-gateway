package main

import (
	"github.com/jukuly/ss_mach_mo/internal/model"
	"github.com/jukuly/ss_mach_mo/internal/model/server"
	"github.com/jukuly/ss_mach_mo/internal/view"
)

func main() {
	var sensors *[]model.Sensor
	model.LoadSensors(model.SENSORS_FILE, sensors)

	go server.StartAdvertising(sensors)
	view.Start(sensors)
}
