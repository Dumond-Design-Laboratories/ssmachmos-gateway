package out

import (
	"log"
	"net"
)

var Logger *log.Logger = log.New(log.Writer(), "", log.Lshortfile|log.LstdFlags)
var PairingConnections map[*net.Conn]bool = make(map[*net.Conn]bool)

func SetLogger(logger *log.Logger) {
	Logger = logger
}

func PairingLog(msg string) {
	for conn := range PairingConnections {
		_, err := (*conn).Write([]byte("MSG:" + msg + "\n"))
		if err != nil {
			Logger.Print(err)
			Logger.Print("Removing connection from PairingConnections")
			delete(PairingConnections, conn)
		}
	}
	Logger.Print(msg)
}
