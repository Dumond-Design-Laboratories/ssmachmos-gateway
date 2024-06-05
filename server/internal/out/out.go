package out

import (
	"log"
	"net"
)

var Logger *log.Logger = log.New(log.Writer(), "", log.LstdFlags)
var PairingConnections map[*net.Conn]bool = make(map[*net.Conn]bool)

func SetLogger(logger *log.Logger) {
	Logger = logger
}

func Error(err error) {
	Logger.Print(err.Error())
}

func Log(msg string) {
	Logger.Print(msg)
}

func PairingLog(msg string) {
	for conn := range PairingConnections {
		_, err := (*conn).Write([]byte("MSG:" + msg + "\n"))
		if err != nil {
			Error(err)
			Log("Removing connection from PairingConnections")
			delete(PairingConnections, conn)
		}
	}
	Log(msg)
}
