package view

import (
	"fmt"

	"github.com/jukuly/ss_mach_mo/internal/model"
)

func DisplaySensors(sensors []model.Sensor) {
	fmt.Print("\n")
	for _, sensor := range sensors {
		fmt.Println(sensor.Name + " - " + sensor.MacToString())
	}
}

func DisplaySensor(sensor model.Sensor) {
	fmt.Print("\n" + sensor.ToString())
}
