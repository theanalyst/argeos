package server

import (
	"context"
	"sync"
	"time"

	"gitlab.cern.ch/eos/argeos/config"
	"gitlab.cern.ch/eos/argeos/internal/common"
	"gitlab.cern.ch/eos/argeos/internal/logger"
	"gitlab.cern.ch/eos/argeos/pkg/plugin"
)

type DiagnosticMonitor struct {
	Cfg               config.ServerConfig
	PluginMgr         *plugin.PluginManager
	interval          time.Duration
	monitoringPlugins []common.HealthDaemon
	healthUpdate      chan common.HealthStatus
}

func NewDiagnosticMonitor(cfg config.ServerConfig, pluginMgr *plugin.PluginManager) *DiagnosticMonitor {
	return &DiagnosticMonitor{
		Cfg:               cfg,
		PluginMgr:         pluginMgr,
		interval:          time.Duration(cfg.DiagnosticInterval) * time.Second,
		monitoringPlugins: make([]common.HealthDaemon, 0),
		healthUpdate:      make(chan common.HealthStatus),
	}
}

func (dm *DiagnosticMonitor) RegisterMonitoringPlugin(plugin common.HealthDaemon) {
	dm.monitoringPlugins = append(dm.monitoringPlugins, plugin)
}

func (dm *DiagnosticMonitor) StartTicker(ctx context.Context) {
	ticker := time.NewTicker(dm.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Logger.Info("Stopping Diagnostic Monitor ticker")
			ticker.Stop()
			return
		case <-ticker.C:
			logger.Logger.Debug("Running periodic health check")
			for _, mp := range dm.monitoringPlugins {
				update := mp.HealthCheck()
				dm.healthUpdate <- update
			}
		}
	}
}

func (dm *DiagnosticMonitor) Start(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	logger.Logger.Info("Starting Diagnostic Monitor")

	for _, plugin := range dm.PluginMgr.Plugins {
		if mp, ok := plugin.(common.HealthDaemon); ok {
			dm.RegisterMonitoringPlugin(mp)
		}
	}

	for _, mp := range dm.monitoringPlugins {
		go func(p common.HealthDaemon) {
			if err := p.Start(ctx, dm.healthUpdate); err != nil {
				logger.Logger.Error("Error starting monitoring plugin", "plugin", p.Name(), "error", err)
			}

		}(mp)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Logger.Info("Stopping Diagnostic Monitor")
				for _, mp := range dm.monitoringPlugins {
					mp.Stop()
				}
				return
			case update := <-dm.healthUpdate:
				logger.Logger.Debug("Received health update", "plugin", update.Name, "status", update.StateString)
				if update.State != common.StateOK {
					dm.PluginMgr.DiagnosticDump(dm.Cfg.DiagnosticDir)
				}
			}
		}
	}()

	dm.StartTicker(ctx)
}
