package server

import (
	"strconv"
	"time"

	"github.com/jukuly/ss_mach_mo/internal/view"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func pairWriteEvent(value []byte, pairingCharacteristic bluetooth.Characteristic, pairing chan bool) {
	if value[0] == 0x00 { // flag => should have one for 1) done pairing 2) pairing request
		// done pairing
		view.Log("Sensor " + string(value) + " has been paired with the Gateway")
		pairingCharacteristic.Write([]byte{})
		<-pairing
	} else if value[0] == 0x01 {
		// pairing request
		view.Log("Sensor " + string(value) + " wants to pair with the Gateway")
		// ask the user if they want to pair
		if true { // condition is temporary until we have a way to ask the user
			pairing <- true // will only allow one sensor to pair at a time (since it's a channel with buffer size 1)
			pairingCharacteristic.Write(value[1:])

			// give 30 second for the sensor to pair, then allow next sensor to pair
			<-time.After(30 * time.Second)
			if len(pairing) > 0 {
				<-pairing
			}
		}
	}
}

func Pair() {
	pairing := make(chan bool, 1)

	var pairingCharacteristic bluetooth.Characteristic
	pairingService := bluetooth.Service{
		UUID: [4]uint32{0x0000FE59, 0x0000FE59, 0x0000FE59, 0x0000FE59}, // same uuid for every gateway
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				Handle: &pairingCharacteristic,
				UUID:   [4]uint32{0x0000FE55, 0x0000FE55, 0x0000FE55, 0x0000FE55}, // same uuid for every gateway
				Value:  []byte{},                                                  // the mac address of the ACCEPTED sensor
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					go pairWriteEvent(value, pairingCharacteristic, pairing)
				}},
		},
	}
	adapter.AddService(&pairingService)
}

func StartAdvertising(updated <-chan bool) {
	adapter.Enable()

	service := bluetooth.Service{
		UUID: [4]uint32{0xA07498CA, 0xAD5B474E, 0x940D16F1, 0xFBE7E8CD},
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID:  [4]uint32{0x51FF12BB, 0x3ED846E5, 0xB4F9D64E, 0x2FEC021B},
				Flags: bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					view.Log("WriteEvent:")
					view.Log("\tOffset: " + strconv.Itoa(offset))
					view.Log("\tValue: " + string(value))
				}},
		}}

	adapter.AddService(&service)

	adv := adapter.DefaultAdvertisement()
	adv.Configure(bluetooth.AdvertisementOptions{
		LocalName: "Gateway Server",
		ServiceUUIDs: []bluetooth.UUID{
			service.UUID,
		}})

	adv.Start()

	view.Log("Advertising started")
}
