package main

import (
	"strconv"
	"time"

	"github.com/jukuly/ss_mach_mo/internal/model"
	"github.com/jukuly/ss_mach_mo/internal/model/server"
	"github.com/jukuly/ss_mach_mo/internal/view"
)

func main() {
	var sensors *[]model.Sensor
	model.LoadSensors(model.SENSORS_FILE, sensors)

	server.Init(sensors)

	go server.StartAdvertising()
	go func() {
		// delay for 15 seconds
		for i := 0; i < 15; i++ {
			view.Log(strconv.Itoa(15 - i))
			time.Sleep(1 * time.Second)
		}

		server.StartPairing()
	}()

	view.Start(sensors)

}
