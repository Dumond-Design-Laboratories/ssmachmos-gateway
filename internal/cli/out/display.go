package out

import (
	"log"
	"strings"

	"github.com/jukuly/ss_mach_mo/internal/model"
)

// var LOG_PATH = "log/"

var Logger *log.Logger = log.New(log.Writer(), "", log.LstdFlags)

func SetLogger(logger *log.Logger) {
	Logger = logger
}

func DisplaySensors(sensors []model.Sensor) {
	if len(sensors) == 0 {
		Logger.Println("No sensors currently paired with the Gateway")
	} else {
		Logger.Print("\n")
		for _, sensor := range sensors {
			Logger.Println(sensor.Name + " - " + model.MacToString(sensor.Mac))
		}
	}
}

func DisplaySensor(sensor model.Sensor) {
	Logger.Print("\n" + sensor.ToString())
}

func Error(err error) {
	Logger.Print(err.Error())
}

func Log(msg string) {
	for _, line := range strings.Split(msg, "\n") {
		Logger.Print(line)
	}
}
