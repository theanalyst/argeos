package server

import (
	"net"

	"gitlab.cern.ch/eos/argeos/internal/logger"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		logger.Logger.Error("Reading from connection ", "error", err)
	}

	message := string(buf[:n])
	logger.Logger.Info("Received", "message", message)
}

func StartServer(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Logger.Error("Starting listener failed", "error", err)
		return
	}

	defer listener.Close()
	logger.Logger.Info("Starting argeos daemon on ", "address", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Logger.Error("Accepting connection ", "error", err)
			continue
		}
		go handleConnection(conn)
	}

}
