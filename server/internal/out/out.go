package out

import (
	"fmt"
	"log"
	"log/syslog"
	"net"
)

type logWriter struct{}

var Logger *log.Logger = log.New(logWriter{}, "", log.Lshortfile|log.LstdFlags)
var sysLogger *log.Logger = nil

var PairingConnections map[*net.Conn]bool = make(map[*net.Conn]bool)
var LoggingConnections map[*net.Conn]bool = make(map[*net.Conn]bool)

func (writer logWriter) Write(bytes []byte) (int, error) {
	// Write to std
	log.Writer().Write(bytes)
	// Write to syslog too
	if sysLogger != nil {
		sysLogger.Writer().Write(bytes)
	}
	return len(bytes), nil
}

func InitSyslog() error {
	var err error = nil
	sysLogger, err = syslog.NewLogger(syslog.LOG_INFO|syslog.LOG_USER, 0)
	return err
}

// Write a MSG command to Unix socket
func Broadcast(msg string) {
	for conn := range LoggingConnections {
		if conn == nil || (*conn) == nil {
			delete(LoggingConnections, conn)
			fmt.Printf("Removing connection %v from LoggingConnections\n", conn)
			continue
		}
		_, err := (*conn).Write([]byte("MSG:" + msg + "\x00"))
		if err != nil {
			delete(LoggingConnections, conn)
			fmt.Println("Error:", err)
			fmt.Printf("Removing connection %v from LoggingConnections\n", conn)
		}
	}
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
