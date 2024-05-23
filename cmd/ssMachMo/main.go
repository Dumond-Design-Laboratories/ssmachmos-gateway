package main

import (
	"github.com/jukuly/ss_mach_mo/internal/model"
	"github.com/jukuly/ss_mach_mo/internal/view/in"
	"github.com/jukuly/ss_mach_mo/internal/view/out"
)

func main() {
	var sensors *[]model.Sensor = &[]model.Sensor{}
	var gateway *model.Gateway = &model.Gateway{}
	model.LoadSensors(model.SENSORS_FILE, sensors)
	err := model.LoadSettings(model.GATEWAY_FILE, gateway)
	if err != nil {
		out.Log("Error loading Gateway settings. Use 'config --id <gateway-id>' and 'config --password <gateway-password>' to set the Gateway settings.")
	}

	//err = server.Init(sensors, gateway)
	//if err != nil {
	//	out.Error(err)
	//	panic("Error initializing server")
	//}

	/*go server.StartAdvertising()
	go func() {
		// delay for 25 seconds
		for i := 0; i < 25; i++ {
			view.Log(strconv.Itoa(25 - i))
			time.Sleep(1 * time.Second)
		}

		server.StartPairing()
	}()*/

	in.Start(sensors, gateway)

}
