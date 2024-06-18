package out

import (
	"log"
	"net"
	"time"
	"fmt"
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
			fmt.Print(err)
			fmt.Printf("Removing connection %v from LoggingConnections\n", conn)
		}
	}
	return len(bytes), nil
}

func SetLogger(logger *log.Logger) {
	Logger = logger
	go func() {
		for {
			time.Sleep(10 * time.Second)
			Logger.Println("still alive")
		}
	}()
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
			fmt.Print(err)
			fmt.Printf("Removing connection %v from PairingConnections\n", conn)
		}
	}
	Logger.Print(msg)
}
