package main

import (
	"github.com/jukuly/ss_mach_mo/internal/model/server"
	"github.com/jukuly/ss_mach_mo/internal/view"
)

func main() {
	go server.StartAdvertising()
	view.Start()
}
