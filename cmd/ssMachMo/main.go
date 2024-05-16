package main

import (
	"github.com/jukuly/ss_mach_mo/internal/model/server"
	"github.com/jukuly/ss_mach_mo/internal/view"
)

func main() {
	updated := make(chan bool)

	go server.StartAdvertising(updated)
	view.Start(updated)
}
