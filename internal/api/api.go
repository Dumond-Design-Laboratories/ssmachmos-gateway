package api

import (
	"net"
	"os"
	"strings"

	"github.com/jukuly/ss_mach_mo/internal/model"
	"github.com/jukuly/ss_mach_mo/internal/model/server"
	"github.com/jukuly/ss_mach_mo/internal/out"
)

func handleCommand(command string, conn *net.Conn) string {
	parts := strings.Split(command, " ")
	if len(parts) == 0 {
		return "ERR:empty command"
	}
	switch parts[0] {
	case "LIST":
		res, err := list()
		if err != nil {
			return "ERR:" + err.Error()
		}
		return "OK:" + res
	case "VIEW":
		if len(parts) < 2 {
			return "ERR:not enough arguments"
		}
		res, err := view(parts[1])
		if err != nil {
			return "ERR:" + err.Error()
		}
		return "OK:" + res
	case "PAIR-ENABLE":
		out.PairingConnections[conn] = true
		err := pairEnable()
		if err != nil {
			return "ERR:" + err.Error()
		}
		return "OK"
	case "PAIR-DISABLE":
		delete(out.PairingConnections, conn)
		if len(out.PairingConnections) == 0 {
			err := pairDisable()
			if err != nil {
				return "ERR:" + err.Error()
			}
			return "OK"
		}
	case "PAIR-ACCEPT":
		if len(parts) < 2 {
			return "ERR:not enough arguments"
		}
		err := pairAccept(parts[1])
		if err != nil {
			return "ERR:" + err.Error()
		}
		return "OK"
	case "FORGET":
		if len(parts) < 2 {
			return "ERR:not enough arguments"
		}
		err := forget(parts[1])
		if err != nil {
			return "ERR:" + err.Error()
		}
		return "OK"
	case "SET-GATEWAY-ID":
		if len(parts) < 2 {
			return "ERR:not enough arguments"
		}
		err := model.SetGatewayId(server.Gateway, parts[1])
		if err != nil {
			return "ERR:" + err.Error()
		}
		return "OK"
	case "SET-GATEWAY-PASSWORD":
		if len(parts) < 2 {
			return "ERR:not enough arguments"
		}
		err := model.SetGatewayPassword(server.Gateway, parts[1])
		if err != nil {
			return "ERR:" + err.Error()
		}
		return "OK"
	case "SET-SENSOR-SETTING":
		return "ERR:unimplemented"
	default:
		return "ERR:invalid command"
	}
	return "ERR:unknown"
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	var buf []byte
	conn.Read(buf)
	response := handleCommand(string(buf), &conn)
	conn.Write([]byte(response))
}

func Start() error {
	socketPath := "/tmp/ss_mach_mos.sock"
	if err := os.RemoveAll(socketPath); err != nil {
		return err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go handleConnection(conn)
	}
}
