package out

import (
	"fmt"
	"time"

	"github.com/jukuly/ss_mach_mo/internal/model"
)

func DisplaySensors(sensors []model.Sensor) {
	fmt.Print("\n")
	for _, sensor := range sensors {
		fmt.Println(sensor.Name + " - " + model.MacToString(sensor.Mac))
	}
}

func DisplaySensor(sensor model.Sensor) {
	fmt.Print("\n" + sensor.ToString())
}

func Error(err error) {
	Log(err.Error())
}

func Log(msg string) {
	fmt.Printf("[%s] %s\n", time.Now().Format(time.RFC3339), msg)
}
