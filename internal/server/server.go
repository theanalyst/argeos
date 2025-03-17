package server

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"gitlab.cern.ch/eos/argeos/config"
	"gitlab.cern.ch/eos/argeos/internal/logger"
	"gitlab.cern.ch/eos/argeos/pkg/plugin"
)

type Server struct {
	Cfg               config.ServerConfig
	PluginMgr         *plugin.PluginManager
	DiagnosticMonitor *DiagnosticMonitor
}

func NewServer(cfg config.ServerConfig, pluginMgr *plugin.PluginManager) *Server {
	return &Server{
		Cfg:               cfg,
		PluginMgr:         pluginMgr,
		DiagnosticMonitor: NewDiagnosticMonitor(cfg, pluginMgr),
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
			cli := scanner.Text()
			parts := strings.Fields(cli)
			if len(parts) == 0 {
				continue
			}

			response := srv.handleCommand(parts[0], parts[1:]...)
			conn.Write([]byte(response + "\n"))
		}
	}
}

func (srv *Server) StartTCPServer(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	address := srv.Cfg.Address
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Logger.Error("Starting TCP listener failed", "error", err)
		return
	}

	defer listener.Close()

	logger.Logger.Info("Starting argeos TCP daemon on ", "address", address)

	go func() {
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
	}()

	<-ctx.Done()
	logger.Logger.Info("TCP Server shutdown complete!")
}

func (srv *Server) StartUnixServer(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
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

	logger.Logger.Info("Starting argeos Unix socket on ", "socketPath", socketPath)

	go func() {
		for {

			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					logger.Logger.Warn("Unix Server shutting down, no longer accepting connection")
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
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-shutdownChan
		logger.Logger.Info("Received Shutdown signal, stopping all services!")
		cancel()
	}()

	wg.Add(1)
	go srv.DiagnosticMonitor.Start(&wg, ctx)

	wg.Add(1)
	go srv.StartUnixServer(&wg, ctx)

	wg.Add(1)
	go srv.StartTCPServer(&wg, ctx)

	wg.Wait()
	logger.Logger.Info("Server shutdown complete!")
}

func (srv *Server) setLogLevel(args ...string) string {
	if len(args) == 0 {
		return "No log level provided, use debug <level>"
	}
	logger.SetLogLevelfromString(args[0])
	return "Log level set to " + args[0]

}

func (srv *Server) handleCommand(command string, args ...string) string {
	switch command {
	case "healthcheck":
		return srv.HealthCheck()
	case "help":
		return srv.PluginMgr.SupportedCommands()
	case "diagnostic_dump":
		return srv.PluginMgr.DiagnosticDump(srv.Cfg.DiagnosticDir)
	case "debug":
		return srv.setLogLevel(args...)
	default:
		return srv.PluginMgr.ExecuteCommand(command, args...)
	}

}

func (srv *Server) HealthCheck() string {
	jsonBytes, err := json.Marshal(srv.PluginMgr.HealthCheck())
	if err != nil {
		logger.Logger.Error("Error json encoding", "err", err)
	}
	return string(jsonBytes)
}
