package gui

import (
	"github.com/jukuly/ss_mach_mo/internal/model"
)

type Console struct {
}

func (c *Console) Write(p []byte) (n int, err error) {

	return len(p), nil
}

func Start(sensors *[]model.Sensor, gateway *model.Gateway) {

}
