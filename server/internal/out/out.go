package out

import (
	"log"
	"net"
)

type logWriter struct{}

var Logger *log.Logger = log.New(logWriter{}, "", log.Lshortfile|log.LstdFlags)
var PairingConnections map[*net.Conn]bool = make(map[*net.Conn]bool)
var LoggingConnections map[*net.Conn]bool = make(map[*net.Conn]bool)

func (writer logWriter) Write(bytes []byte) (int, error) {
	log.Writer().Write(bytes)
	for conn := range LoggingConnections {
		_, err := (*conn).Write([]byte("LOG:" + string(bytes) + "\x00"))
		if err != nil {
			delete(LoggingConnections, conn)
			Logger.Print(err)
			Logger.Printf("Removing connection %v from LoggingConnections", conn)
		}
	}
	return len(bytes), nil
}

func SetLogger(logger *log.Logger) {
	Logger = logger
}

func PairingLog(msg string) {
	for conn := range PairingConnections {
		_, err := (*conn).Write([]byte("MSG:" + msg + "\x00"))
		if err != nil {
			delete(PairingConnections, conn)
			Logger.Print(err)
			Logger.Printf("Removing connection %v from PairingConnections", conn)
		}
	}
	Logger.Print(msg)
}
