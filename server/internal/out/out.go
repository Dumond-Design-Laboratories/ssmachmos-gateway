package out

import (
	"fmt"
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
		if conn == nil || (*conn) == nil {
			delete(LoggingConnections, conn)
			fmt.Printf("Removing connection %v from LoggingConnections\n", conn)
			continue
		}
		_, err := (*conn).Write([]byte("LOG:" + string(bytes) + "\x00"))
		if err != nil {
			delete(LoggingConnections, conn)
			fmt.Println("Error:", err)
			fmt.Printf("Removing connection %v from LoggingConnections\n", conn)
		}
	}
	return len(bytes), nil
}

func SetLogger(logger *log.Logger) {
	Logger = logger
}

func Error(content interface{}) {
	Logger.Println("Error:", content)
}

func PairingLog(msg string) {
	for conn := range PairingConnections {
		if conn == nil || (*conn) == nil {
			delete(PairingConnections, conn)
			fmt.Printf("Removing connection %v from PairingConnections\n", conn)
			continue
		}
		_, err := (*conn).Write([]byte("MSG:" + msg + "\x00"))
		if err != nil {
			delete(PairingConnections, conn)
			fmt.Println("Error:", err)
			fmt.Printf("Removing connection %v from PairingConnections\n", conn)
		}
	}
	Logger.Println(msg)
}
