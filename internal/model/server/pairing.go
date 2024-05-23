package server

import (
	"time"

	"github.com/jukuly/ss_mach_mo/internal/model"
	"github.com/jukuly/ss_mach_mo/internal/view/out"
)

type pairingRequest struct {
	mac        [6]byte
	expiration time.Time
}

type PairingState struct {
	active    bool
	requested []pairingRequest
	pairing   [6]byte
}

var pairingState = PairingState{}

func StartPairing() {
	pairingState.active = true
	out.Log("Pairing started")
}

func StopPairing() {
	pairingState.active = false
	out.Log("Pairing stopped")
}

func pairRequest(value []byte) {
	if len(value) != 6 || !pairingState.active {
		return
	}
	mac := [6]byte(value[:6])
	if pairingState.pairing == mac {
		return
	}
	existed := false
	for _, p := range pairingState.requested {
		if p.mac == mac {
			if time.Now().Before(p.expiration) {
				return
			} else {
				p.expiration = time.Now().Add(30 * time.Second)
				existed = true
				break
			}
		}
	}
	if !existed {
		pairingState.requested = append(pairingState.requested, pairingRequest{
			mac:        mac,
			expiration: time.Now().Add(30 * time.Second),
		})
	}

	out.Log("Pair request from " + model.MacToString(mac) + " | pair --accept <mac-address> to accept")
}

// TODO add data characteristic UUID to confirmation
func pairConfirmation(value []byte) {
	if len(value) != 6 || !pairingState.active {
		return
	}
	mac := [6]byte(value[:6])
	if pairingState.pairing != mac {
		return
	}
	pairingState.pairing = [6]byte{}
	pairResponseCharacteristic.Write([]byte{})

	out.Log(model.MacToString(mac) + " has been paired with the Gateway")
}

// TODO add data characteristic UUID to response
func Pair(mac [6]byte) {
	if !pairingState.active {
		out.Log("Pairing is not active")
		return
	}
	found := false
	for _, p := range pairingState.requested {
		if p.mac == mac && time.Now().Before(p.expiration) {
			found = true
			break
		}
	}
	if !found {
		out.Log("Pair request from " + model.MacToString(mac) + " not found")
		return
	}
	if pairingState.pairing != [6]byte{} {
		out.Log("Canceled pairing with " + model.MacToString(pairingState.pairing))
	}
	pairingState.pairing = mac

	pairResponseCharacteristic.Write(mac[:])
	out.Log("Pairing with " + model.MacToString(mac))

	go func() {
		time.Sleep(30 * time.Second)
		if pairingState.pairing == mac {
			pairingState.pairing = [6]byte{}
			pairResponseCharacteristic.Write([]byte{})
			out.Log("Pairing with " + model.MacToString(mac) + " has timed out")
		}
	}()
}
