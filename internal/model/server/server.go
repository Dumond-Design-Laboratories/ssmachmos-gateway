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

	service := bluetooth.Service{
		UUID: [4]uint32{0xA07498CA, 0xAD5B474E, 0x940D16F1, 0xFBE7E8CD},
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID:  [4]uint32{0x51FF12BB, 0x3ED846E5, 0xB4F9D64E, 0x2FEC021B},
				Value: []byte("Hello World!"),
				Flags: bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					fmt.Println("WriteEvent:")
					fmt.Printf("\toffset: %v\n", offset)
					fmt.Printf("\tvalue: %v\n", value)
				}},
		}}

	adapter.AddService(&service)

	adv := adapter.DefaultAdvertisement()
	adv.Configure(bluetooth.AdvertisementOptions{
		LocalName: "BLE Test Go",
		ServiceUUIDs: []bluetooth.UUID{
			service.UUID,
		}})

	adv.Start()

	fmt.Println("Advertising started")
}
