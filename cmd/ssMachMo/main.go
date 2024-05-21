package main

import (
	"github.com/jukuly/ss_mach_mo/internal/model"
	"github.com/jukuly/ss_mach_mo/internal/model/server"
	"github.com/jukuly/ss_mach_mo/internal/view"
)

func main() {
	var sensors *[]model.Sensor = &[]model.Sensor{}
	model.LoadSensors(model.SENSORS_FILE, sensors)

	server.InitAlt(sensors)

	/*go server.StartAdvertising()
	go func() {
		// delay for 25 seconds
		for i := 0; i < 25; i++ {
			view.Log(strconv.Itoa(25 - i))
			time.Sleep(1 * time.Second)
		}

		server.StartPairing()
	}()*/

	view.Start(sensors)

}
