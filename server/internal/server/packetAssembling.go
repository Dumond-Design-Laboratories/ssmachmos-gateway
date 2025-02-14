package server

import (
	"encoding/binary"
	"time"

	"github.com/jukuly/ss_machmos/server/internal/model"
	"github.com/jukuly/ss_machmos/server/internal/out"
)

type Packet struct {
	offset int
	data   []byte
}

type Transmission struct {
	macAddress        [6]byte
	batteryLevel      int
	timestamp         string
	dataType          string
	samplingFrequency uint32
	currentLength     int
	totalLength       uint32
	packets           []byte
}

// https://go.dev/doc/faq#atomic_maps
// :^)
var transmissions map[[6]byte]Transmission

// Collect all packets into a single array after finishing transmit and return
// func (t *Transmission) AssemblePackets() []byte {
// 	data := make([]byte, transmission.totalLength)
// 	for _, packet := range transmission.packets {
// 		if packet.offset+len(packet.data) > len(data) {
// 			continue
// 		}
// 		data = append(append(data[:packet.offset], packet.data...), data[packet.offset+len(packet.data):]...)
// 	}
// 	return data
// }

// Grab packet from network, save to memory
// We're just recording an octet stream, of which the first ever packet should be a short header
// The rest is byte soup
// Grab header from initial packet
// batteryLevel := int(int8(data[7]))
// dataType := DATA_TYPES[data[8]]
// samplingFrequency := binary.LittleEndian.Uint32(data[9:13])
// lengthOfData := binary.LittleEndian.Uint32(data[13:17])
// messageID := *(*[3]byte)(data[17:20])
// offset := int(binary.LittleEndian.Uint32(data[20:24]))
func savePacket(data []byte, macAddress [6]byte, dataType string) (t Transmission, ok bool) {
	if _, exists := transmissions[macAddress]; !exists {
		// New transmission
		out.Logger.Println("COLLECT-START:" + model.MacToString(macAddress))
		// First packet is a header
		totalLength := binary.LittleEndian.Uint32(data[0:4])
		samplingFrequency := binary.LittleEndian.Uint32(data[4:8])
		batteryLevel := -1
		transmissions[macAddress] = Transmission{
			macAddress:        macAddress,
			timestamp:         time.Now().UTC().Format(ISO8601),
			batteryLevel:      batteryLevel,
			dataType:          dataType,
			samplingFrequency: samplingFrequency,
			currentLength:     0,
			totalLength:       totalLength,
			packets:           make([]byte, totalLength),
		}
	} else {
		// Other packets are raw data
		transmission := transmissions[macAddress]
		// increase current byte count
		transmission.currentLength += len(data)
		// Append data to end of stream
		transmission.packets = append(transmission.packets, data...)
		transmissions[macAddress] = transmission
		out.Logger.Println("Received packet from", model.MacToString(macAddress), "total", transmission.currentLength, "/", transmission.totalLength)
	}

	// Header includes expected length, expect more from that
	if transmissions[macAddress].currentLength >= int(transmissions[macAddress].totalLength) {
		//fullRawData := assemblePackets(transmissions[macAddress])
		fullTransmit := transmissions[macAddress]
		delete(transmissions, macAddress)
		out.Logger.Println("Assembled transmission packets for " + model.MacToString(macAddress))
		out.Logger.Println("COLLECT-END:" + model.MacToString(macAddress))
		return fullTransmit, true
	}

	// Nothing to return
	return Transmission{}, false
}
