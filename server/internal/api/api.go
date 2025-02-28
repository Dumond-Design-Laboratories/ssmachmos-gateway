package api

import (
	"bufio"
	"net"
	"os"
	"slices"
	"strings"

	"github.com/jukuly/ss_machmos/server/internal/model"
	"github.com/jukuly/ss_machmos/server/internal/out"
	"github.com/jukuly/ss_machmos/server/internal/server"
)

var connectionsAlive []*net.Conn;

func handleCommand(command string, conn *net.Conn) string {
	parts := strings.Split(command, " ")

	if len(parts) == 0 {
		return "ERR:empty command"
	}
	switch parts[0] {
	case "LIST":
		// List devices paired
		res, err := list()
		if err != nil {
			out.Logger.Println("Error:", err)
			return "ERR:LIST:" + err.Error()
		}
		return "OK:LIST:" + res
	case "LIST-CONNECTED":
		// List device connection status
		res, err := listConnected()
		if err != nil {
			out.Logger.Println("Error:", err)
			return "ERR:LIST-CONNECTED:" + err.Error()
		}
		return "OK:LIST-CONNECTED:"+res;
	case "COLLECT":
		if len(parts) < 2 {
			return "ERR:COLLECT:Not enough arguments, missing mac address"
		}
		err := deviceCollect(parts[1])
		if err != nil {
			out.Logger.Println("Error: ", err)
			return "ERR:COLLECT:" + err.Error()
		}
		return "OK:COLLECT:"
	case "LIST-PENDING-UPLOADS":
		res, err := pendingUploads();
		if err != nil {
			out.Logger.Println("Error:", err)
			return "ERR:LIST-PENDING-UPLOADS:"+err.Error()
		}
		return "OK:LIST-PENDING-UPLOADS:"+res;
	case "VIEW":
		if len(parts) < 2 {
			return "ERR:not enough arguments"
		}
		res, err := view(parts[1])
		if err != nil {
			out.Logger.Println("Error:", err)
			return "ERR:VIEW:" + err.Error()
		}
		return "OK:VIEW:" + res
	case "PAIR-LIST":
		res, err := pairListPending()
		if err != nil {
			out.Logger.Println("Error:", err)
			return "ERR:PAIR-LIST:" + err.Error()
		}
		return "OK:PAIR-LIST:" + res
	case "PAIR-ENABLE":
		// Save this connection as someone interested in pairing info
		out.PairingConnections[conn] = true
		pairEnable()
		return "OK:PAIR-ENABLE:"
	case "PAIR-DISABLE":
		// Remove connection from update list
		delete(out.PairingConnections, conn)
		if len(out.PairingConnections) == 0 {
			pairDisable()
		}
		return "OK:PAIR-DISABLE:"
	case "PAIR-ACCEPT":
		if len(parts) < 2 {
			return "ERR:PAIR-ACCEPT:not enough arguments"
		}
		err := pairAccept(parts[1])
		if err != nil {
			out.Logger.Println("Error:", err)
			return "ERR:PAIR-ACCEPT:" + err.Error()
		}
		return "OK:PAIR-ACCEPT:"
	case "FORGET":
		if len(parts) < 2 {
			return "ERR:FORGET:not enough arguments"
		}
		err := forget(parts[1])
		if err != nil {
			out.Logger.Println("Error:", err)
			return "ERR:FORGET:" + err.Error()
		}
		return "OK:FORGET:"
	case "GET-GATEWAY":
		res, err := getGateway()
		if err != nil {
			out.Logger.Println("Error:", err)
			return "ERR:GET-GATEWAY:" + err.Error()
		}
		return "OK:GET-GATEWAY:" + res
	case "SET-GATEWAY-HTTP-ENDPOINT":
		if len(parts) < 2 {
			return "ERR:SET-GATEWAY-HTTP-ENDPOINT:not enough arguments"
		}
		var err error
		if parts[1] == "default" {
			err = model.SetGatewayHTTPEndpoint(server.Gateway, server.DEFAULT_GATEWAY_HTTP_ENDPOINT)
		} else {
			err = model.SetGatewayHTTPEndpoint(server.Gateway, parts[1])
		}
		if err != nil {
			out.Logger.Println("Error:", err)
			return "ERR:SET-GATEWAY-HTTP-ENDPOINT:" + err.Error()
		}
		return "OK:SET-GATEWAY-HTTP-ENDPOINT:"
	case "SET-GATEWAY-ID":
		if len(parts) < 2 {
			return "ERR:SET-GATEWAY-ID:not enough arguments"
		}
		err := model.SetGatewayId(server.Gateway, strings.Join(parts[1:], " "))
		if err != nil {
			out.Logger.Println("Error:", err)
			return "ERR:SET-GATEWAY-ID:" + err.Error()
		}
		return "OK:SET-GATEWAY-ID:"
	case "SET-GATEWAY-PASSWORD":
		if len(parts) < 2 {
			return "ERR:not enough arguments"
		}
		err := model.SetGatewayPassword(server.Gateway, strings.Join(parts[1:], " "))
		if err != nil {
			out.Logger.Println("Error:", err)
			return "ERR:SET-GATEWAY-PASSWORD:" + err.Error()
		}
		return "OK:SET-GATEWAY-PASSWORD:"
	case "SET-SENSOR-SETTINGS":
		if len(parts) < 2 {
			return "ERR:SET-SENSOR-SETTINGS:not enough arguments"
		}
		mac, err := model.StringToMac(parts[1])
		if err != nil {
			out.Logger.Println("Error:", err)
			return "ERR:SET-SENSOR-SETTING:" + err.Error()
		}
		// Parts is split by spaces
		// For each part we call UpdateSensorSetting
		nbrOfSettings := (len(parts) - 2) / 2
		for i := 0; i < nbrOfSettings; i++ {
			// FIXME: Replace this with JSON instead
			err = model.UpdateSensorSetting(mac, parts[2+i*2], parts[3+i*2], server.Sensors)
			if err != nil {
				out.Logger.Println("Error:", err)
				return "ERR:SET-SENSOR-SETTINGS:" + err.Error()
			}
		}
		server.TriggerSettingCollection();
		return "OK:SET-SENSOR-SETTINGS:"
	case "ADD-LOGGER":
		out.LoggingConnections[conn] = true
		out.Logger.Println("Adding logger");
		return "OK:ADD-LOGGER:"
	case "REMOVE-LOGGER":
		delete(out.LoggingConnections, conn)
		return "OK:REMOVE-LOGGER:"
	case "STOP":
		stop()
	default:
		return "ERR:invalid command " + command
	}
	return "ERR:unknown"
}

// Handle command and write response
func handleConnection(conn *net.Conn) {
	defer (*conn).Close()

	connectionsAlive = append(connectionsAlive, conn)
	defer slices.Delete(connectionsAlive,
		slices.Index(connectionsAlive, conn),
		slices.Index(connectionsAlive, conn)+1)

	reader := bufio.NewReader(*conn)
	for {
		str, err := reader.ReadString('\x00')
		if err != nil {
			return
		}
		cs := strings.Split(str, "\x00")
		for _, c := range cs {
			if c == "" {
				continue
			}
			response := handleCommand(c, conn)
			// Terminate with zero byte
			// For socat you need to insert the zero byte character to terminate
			// echo -e "PAIR-LIST\0"
			// or printf "PAIR-LIST\0"
			// or CTRL-V then CTRL-SHIFT-2 to insert ^@ and terminate command
			(*conn).Write([]byte(response + "\x00"))
		}
	}
}

// Send a message to all alive connections
// func Broadcast(msg string) {
// 	for _, conn := range connectionsAlive {
// 		(*conn).Write([]byte(msg + "\x00"))
// 	}
// 	out.Logger.Println(msg)
// }

// Start listening to unix socket
// any commands received here would be sent to handleConnection and then handleCommand
// Commands are zero terminated. Unix sockets are bidirectional
// https://beej.us/guide/bgipc/html/index-wide.html#unixsock
func Start() error {
	socketPath := "/tmp/ss_machmos.sock"
	if err := os.RemoveAll(socketPath); err != nil {
		out.Logger.Println("Error:", err)
		return err
	}

	// Setup receiver
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		out.Logger.Println("Error:", err)
		return err
	}
	defer listener.Close()

	for {
		// Accept returns another socket descriptor
		conn, err := listener.Accept()
		if err != nil {
			out.Logger.Println("Error:", err)
			return err
		}
		go handleConnection(&conn)
	}
}
