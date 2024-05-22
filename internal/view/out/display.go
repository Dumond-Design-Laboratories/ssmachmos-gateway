package out

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jukuly/ss_mach_mo/internal/model"
)

var LOG_PATH = "log/"

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
	str := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), msg)
	fmt.Print(str)

	_, err := os.Stat(LOG_PATH)
	if os.IsNotExist(err) {
		os.MkdirAll(LOG_PATH, os.ModePerm)
	}

	path, _ := filepath.Abs(LOG_PATH + time.Now().Format(time.DateOnly) + ".txt")

	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("[%s] %s\n", time.Now().Format(time.RFC3339), err.Error())
	}
	logFile.WriteString(str)
}
