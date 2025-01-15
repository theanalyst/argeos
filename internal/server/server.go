package server

import (
	"bufio"
	"context"
	"encoding/json"
	"gitlab.cern.ch/eos/argeos/internal/config"
	"gitlab.cern.ch/eos/argeos/internal/logger"
	"gitlab.cern.ch/eos/argeos/pkg/plugin"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	Cfg       config.ServerConfig
	PluginMgr *plugin.PluginManager
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

func (srv *Server) handleConnectionWithCtx(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	select {
	case <-ctx.Done():
		logger.Logger.Info("Stopping Connection Handler due to server shutdown")
		return
	default:
		scanner := bufio.NewScanner(conn)

		for scanner.Scan() {
			command := scanner.Text()
			response := srv.handleCommand(command)
			conn.Write([]byte(response + "\n"))
		}
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-shutdownChan
		logger.Logger.Info("Received Shutdown signal, stopping TCP server")
		cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			logger.Logger.Warn("TCP Server shutting down, no longer accepting connection")
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					logger.Logger.Error("Accepting connection ", "error", err)
					continue
				}

			}
			go srv.handleConnectionWithCtx(ctx, conn)

		}
	}
}

func (srv *Server) StartUnixServer() {
	var socketPath = srv.Cfg.AdminSocket
	listener, err := net.Listen("unix", socketPath)

	if err != nil {
		logger.Logger.Error("Starting Unix listener failed", "error", err)
		return
	}

	defer func() {
		if err := listener.Close(); err != nil {
			logger.Logger.Error("Error closing socket", "error", err)
		}

		if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
			logger.Logger.Error("Failed to remove socket", "error", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-shutdownChan
		logger.Logger.Info("Received Shutdown signal, stopping UDP server")
		cancel()
	}()

	logger.Logger.Info("Starting argeos Unix socket on ", "socketPath", socketPath)

	go func() {
		for {

			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					logger.Logger.Error("Accepting connection", "error", err)
				}
				continue
			}
			go srv.handleConnectionWithCtx(ctx, conn)

		}
	}()

	<-ctx.Done()
	logger.Logger.Info("Unix Server shutdown complete!")
}

func (srv *Server) Start() {
	srv.StartUnixServer()
	srv.StartTCPServer()
}

func (srv *Server) handleCommand(command string) string {
	switch command {
	case "healthcheck":
		return srv.HealthCheck()
	case "help":
		fallthrough
	default:
		return "Unknown command"
	}

}

func (srv *Server) HealthCheck() string {
	jsonBytes, err := json.Marshal(srv.PluginMgr.HealthCheck())
	if err != nil {
		logger.Logger.Error("Error json encoding", err)
	}
	return string(jsonBytes)
}
