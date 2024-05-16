package server

import (
	"time"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func Pair() {
}

func StartAdvertising(updated <-chan bool) {
	adapter.Enable()

	// Define the peripheral device info.
	adv := adapter.DefaultAdvertisement()
	adv.Configure(bluetooth.AdvertisementOptions{
		LocalName: "Go Bluetooth",
	})

	// Start advertising
	adv.Start()

	println("advertising...")
	for {
		// Sleep forever.
		time.Sleep(time.Hour)
	}
}
