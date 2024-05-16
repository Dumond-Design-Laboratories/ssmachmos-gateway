package server

import (
	"fmt"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func Pair() {
}

func StartAdvertising(updated <-chan bool) {
	adapter.Enable()

	adapter.AddService(&bluetooth.Service{
		UUID: [4]uint32{0x5F9B34FB, 0x80000080, 0x00001000, 0x1800},
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID:  [4]uint32{0x5F9B34FB, 0x80000080, 0x00001000, 0x2A00},
				Value: []byte("Hello World!"),
				Flags: bluetooth.CharacteristicReadPermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					fmt.Println("WriteEvent:")
					fmt.Printf("\toffset: %v\n", offset)
					fmt.Printf("\tvalue: %v\n", value)
				}},
		}})

	adv := adapter.DefaultAdvertisement()
	adv.Configure(bluetooth.AdvertisementOptions{
		LocalName: "BLE Test Go"})

	adv.Start()

	fmt.Println("Advertising started")
}
