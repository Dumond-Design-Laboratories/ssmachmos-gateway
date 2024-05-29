package out

import (
	"fmt"
	"log"
	"time"

	"github.com/jukuly/ss_mach_mo/internal/model"
)

// var LOG_PATH = "log/"

var Logger *log.Logger = log.Default()

func SetLogger(logger *log.Logger) {
	Logger = logger
}

func DisplaySensors(sensors []model.Sensor) {
	for _, sensor := range sensors {
		Logger.Println(sensor.Name + " - " + model.MacToString(sensor.Mac))
	}
}

func DisplaySensor(sensor model.Sensor) {
	Logger.Print(sensor.ToString())
}

func Error(err error) {
	Log(err.Error())
}

func Log(msg string) {
	str := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), msg)
	Logger.Print(str)

	/*_, err := os.Stat(LOG_PATH)
	if os.IsNotExist(err) {
		os.MkdirAll(LOG_PATH, os.ModePerm)
	}

	path, _ := filepath.Abs(LOG_PATH + time.Now().Format(time.DateOnly) + ".txt")

	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		Logger.Print("[%s] %s\n", time.Now().Format(time.RFC3339), err.Error())
	}
	logFile.WriteString(str)*/
}
