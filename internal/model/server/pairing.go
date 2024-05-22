package server

import (
	"time"

	"github.com/jukuly/ss_mach_mo/internal/model"
	"github.com/jukuly/ss_mach_mo/internal/view/out"
	"tinygo.org/x/bluetooth"
)

func StartPairing(state *PairingState) {
	state.active = true
	out.Log("Pairing started")
}

func StopPairing(state *PairingState) {
	state.active = false
	out.Log("Pairing stopped")
}

func pairRequest(value []byte, state *PairingState) {
	if len(value) != 6 || !state.active {
		return
	}
	mac := [6]byte(value[:6])
	if state.pairing == mac {
		return
	}
	existed := false
	for _, p := range state.requested {
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
		state.requested = append(state.requested, pairingRequest{
			mac:        mac,
			expiration: time.Now().Add(30 * time.Second),
		})
	}

	out.Log("Pair request from " + model.MacToString(mac) + " | pair <mac-address> to accept")
}

func pairConfirmation(value []byte, pairResponse bluetooth.Characteristic, state *PairingState) {
	if len(value) != 6 || !state.active {
		return
	}
	mac := [6]byte(value[:6])
	if state.pairing != mac {
		return
	}
	state.pairing = [6]byte{}

	pairResponse.Write([]byte{})
	out.Log(model.MacToString(mac) + " has been paired with the Gateway")
}

func Pair(mac [6]byte, pairResponse bluetooth.Characteristic, state *PairingState) {
	if !state.active {
		out.Log("Pairing is not active")
		return
	}
	found := false
	for _, p := range state.requested {
		if p.mac == mac && time.Now().Before(p.expiration) {
			found = true
			break
		}
	}
	if !found {
		out.Log("Pair request from " + model.MacToString(mac) + " not found")
		return
	}
	if state.pairing != [6]byte{} {
		out.Log("Canceled pairing with " + model.MacToString(state.pairing))
		return
	}
	state.pairing = mac

	pairResponse.Write(mac[:])
	out.Log("Pairing with " + model.MacToString(mac))

	go func() {
		time.Sleep(30 * time.Second)
		if state.pairing == mac {
			state.pairing = [6]byte{}
			out.Log("Pairing with " + model.MacToString(mac) + " has timed out")
		}
	}()
}
