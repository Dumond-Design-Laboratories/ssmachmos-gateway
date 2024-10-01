package server

type Packet struct {
	offset int
	data   []byte
}

type Transmission struct {
	macAddress        [6]byte
	batteryLevel      int
	dataType          string
	samplingFrequency uint32
	currentLength     int
	totalLength       uint32
	packets           []Packet
}

var transmissions map[[3]byte]Transmission

func assemblePackets(transmission Transmission) []byte {
	data := make([]byte, transmission.totalLength)
	for _, packet := range transmission.packets {
		if packet.offset+len(packet.data) > len(data) {
			continue
		}
		data = append(append(data[:packet.offset], packet.data...), data[packet.offset+len(packet.data):]...)
	}
	return data
}

func savePacket(data []byte, macAddress [6]byte, batteryLevel int, dataType string, samplingFrequency uint32, totalLength uint32, messageID [3]byte, offset int) []byte {
	if _, exists := transmissions[messageID]; !exists {
		transmissions[messageID] = Transmission{
			macAddress:        macAddress,
			batteryLevel:      batteryLevel,
			dataType:          dataType,
			samplingFrequency: samplingFrequency,
			currentLength:     len(data),
			totalLength:       totalLength,
			packets: []Packet{
				{
					offset: offset,
					data:   data,
				},
			},
		}
	} else {
		transmission := transmissions[messageID]
		transmission.currentLength += len(data)
		transmission.packets = append(transmission.packets, Packet{
			offset: offset,
			data:   data,
		})
		transmissions[messageID] = transmission
	}

	if transmissions[messageID].currentLength >= int(transmissions[messageID].totalLength) {
		fullRawData := assemblePackets(transmissions[messageID])
		delete(transmissions, messageID)
		return fullRawData
	}

	return nil
}
