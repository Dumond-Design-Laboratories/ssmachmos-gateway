package server

import (
	"crypto/rsa"
	"encoding/binary"
	"time"

	"github.com/jukuly/ss_machmos/server/internal/model"
	"github.com/jukuly/ss_machmos/server/internal/out"
)

type request struct {
	publicKey          *rsa.PublicKey
	dataTypes          []string
	collectionCapacity uint32
}

type pairingState struct {
	active    bool
	requested map[[6]byte]request
	pairing   [6]byte
}

var state pairingState

func EnablePairing() {
	state.active = true
}

func DisablePairing() {
	state.active = false
}

// see protocol.md to understand what is going on here
func pairRequest(value []byte) {
	if len(value) < 11 || !state.active {
		return
	}
	mac := [6]byte(value[:6])
	for _, s := range *Sensors {
		if s.Mac == mac {
			out.PairingLog("REQUEST-SENSOR-EXISTS:" + model.MacToString(mac))
			return
		}
	}
	if _, exists := state.requested[mac]; exists {
		return
	}

	dataTypes := []string{}
	if value[6]&0x01 == 0x01 {
		dataTypes = append(dataTypes, "acoustic")
	}
	if value[6]&0x02 == 0x02 {
		dataTypes = append(dataTypes, "temperature")
	}
	if value[6]&0x04 == 0x04 {
		dataTypes = append(dataTypes, "vibration")
	}

	collectionCapacity := binary.LittleEndian.Uint32(value[7:11])
	publicKey, err := model.ParsePublicKey(value[11:])
	if err != nil {
		return
	}

	state.requested[mac] = request{
		publicKey:          publicKey,
		dataTypes:          dataTypes,
		collectionCapacity: collectionCapacity,
	}

	go func() {
		time.Sleep(30 * time.Second)
		if _, exists := state.requested[mac]; exists && state.pairing != mac {
			delete(state.requested, mac)
			out.PairingLog("REQUEST-TIMEOUT:" + model.MacToString(mac))
		}
	}()

	out.PairingLog("REQUEST-NEW:" + model.MacToString(mac))
}

// see protocol.md to understand what is going on here
func pairConfirmation(value []byte) {
	if len(value) != 294 || !state.active {
		return
	}

	data := value[:38]
	mac := [6]byte(data[:6])
	dataUuid := model.BytesToUuid([16]byte(data[6:22]))
	settingsUuid := model.BytesToUuid([16]byte(data[22:38]))
	signature := value[len(value)-256:]

	if _, exists := state.requested[mac]; !exists || state.pairing != mac || !model.VerifySignature(data, signature, state.requested[mac].publicKey) {
		return
	}

	dataCharUUID, err := model.GetDataCharUUID(Gateway)
	if err != nil || dataCharUUID != dataUuid {
		return
	}
	settingsCharUUID, err := model.GetSettingsCharUUID(Gateway)
	if err != nil || settingsCharUUID != settingsUuid {
		return
	}
	state.pairing = [6]byte{}
	pairResponseCharacteristic.Write([]byte{})
	model.AddSensor(mac, state.requested[mac].dataTypes, state.requested[mac].collectionCapacity, state.requested[mac].publicKey, Sensors)
	delete(state.requested, mac)

	out.PairingLog("PAIR-SUCCESS:" + model.MacToString(mac))
}

// see protocol.md to understand what is going on here
func Pair(mac [6]byte) {
	if !state.active {
		out.PairingLog("PAIRING-DISABLED")
		return
	}

	if _, exists := state.requested[mac]; !exists {
		out.PairingLog("REQUEST-NOT-FOUND:" + model.MacToString(mac))
		return
	}

	if state.pairing != [6]byte{} && state.pairing != mac {
		out.PairingLog("PAIRING-CANCELED:" + model.MacToString(state.pairing))
		delete(state.requested, mac)
	}
	state.pairing = mac

	dataCharUUID, _ := model.GetDataCharUUID(Gateway)
	settingsCharUUID, _ := model.GetSettingsCharUUID(Gateway)
	dataUuid := model.UuidToBytes(dataCharUUID)
	settingsUuid := model.UuidToBytes(settingsCharUUID)
	pairResponseCharacteristic.Write(append(append(append(mac[:], 0x01), dataUuid[:]...), settingsUuid[:]...))
	out.PairingLog("PAIRING-WITH:" + model.MacToString(mac))

	go func() {
		time.Sleep(30 * time.Second)
		if state.pairing == mac {
			state.pairing = [6]byte{}
			pairResponseCharacteristic.Write([]byte{})
			delete(state.requested, mac)
			out.PairingLog("PAIRING-TIMEOUT:" + model.MacToString(mac))
		}
	}()
}
