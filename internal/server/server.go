package server

import (
	"net"
	"bufio"
	"gitlab.cern.ch/eos/argeos/internal/logger"
	"gitlab.cern.ch/eos/argeos/internal/config"
)


type Server struct {
	Cfg config.ServerConfig
}

func (srv *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		command := scanner.Text()
		response := srv.handleCommand(command)
		conn.Write([]byte(response + "\n"))
	}

}

func (srv *Server) StartTCPServer() {
	address := srv.Cfg.Address
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Logger.Error("Starting TCP listener failed", "error", err)
		return
	}

	defer listener.Close()
	logger.Logger.Info("Starting argeos TCP daemon on ", "address", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Logger.Error("Accepting connection ", "error", err)
			continue
		}
		go srv.handleConnection(conn)
	}

}


func (srv *Server) StartUnixServer() {
	var socketPath = srv.Cfg.AdminSocket
	listener, err := net.Listen("unix", socketPath)

	if err != nil {
		logger.Logger.Error("Starting Unix listener failed", "error", err)
		return
	}

	defer listener.Close()
	logger.Logger.Info("Starting argeos Unix socket on ", socketPath)

	for {

		conn, err := listener.Accept()
		if err != nil {
			logger.Logger.Error("Accepting connection", "error", err)
			continue
		}
		go srv.handleConnection(conn)
	}



}

func (srv *Server) Start() {
	srv.StartUnixServer()
	srv.StartTCPServer()
}


func (srv *Server) handleCommand(command string) string {
	switch(command) {
	case "healthcheck":
		return srv.HealthCheck()
	case "help":
		fallthrough
	default:
		return "Unknown command"
	}

}


func (srv *Server) HealthCheck() string {
	return "HEALTH_OK"
}
