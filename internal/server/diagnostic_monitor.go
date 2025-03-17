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
	consecutiveFails  int
	backOffDuration   time.Duration
	maxBackOff        time.Duration
}

func NewDiagnosticMonitor(cfg config.ServerConfig, pluginMgr *plugin.PluginManager) *DiagnosticMonitor {
	return &DiagnosticMonitor{
		Cfg:               cfg,
		PluginMgr:         pluginMgr,
		interval:          time.Duration(cfg.DiagnosticInterval) * time.Second,
		monitoringPlugins: make([]common.HealthDaemon, 0),
		healthUpdate:      make(chan common.HealthStatus, 100),
		consecutiveFails:  0,
		backOffDuration:   1 * time.Second,
		maxBackOff:        1800 * time.Second,
	}
}

func (dm *DiagnosticMonitor) RegisterMonitoringPlugin(plugin common.HealthDaemon) {
	logger.Logger.Info("Registering monitoring plugin", "plugin", plugin.Name())
	dm.monitoringPlugins = append(dm.monitoringPlugins, plugin)
}

func (dm *DiagnosticMonitor) StartTicker(ctx context.Context) {
	logger.Logger.Info("Starting Diagnostic Monitor ticker")
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
		} else {
			logger.Logger.Debug("Plugin does not implement HealthDaemon interface", "plugin", plugin.Name())
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
		var backoffTimer *time.Timer
		isBackingOff := false

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
				if update.State == common.StateFAIL {
					dm.consecutiveFails++
					if !isBackingOff {
						isBackingOff = true
						delay := min(dm.backOffDuration*time.Duration(dm.consecutiveFails), dm.maxBackOff)
						backoffTimer = time.NewTimer(delay)
						logger.Logger.Warn("Health check failed", "plugin", update.Name, "consecutiveFails", dm.consecutiveFails, "backoff", delay)
					}
				} else if update.State == common.StateOK {
					dm.consecutiveFails = 0
					isBackingOff = false
					if backoffTimer != nil && !backoffTimer.Stop() {
						<-backoffTimer.C // Drain the channel
					}
					backoffTimer = nil
				}
			case <-func() <-chan time.Time {
				if backoffTimer != nil {
					return backoffTimer.C
				}
				return nil
			}():
				if isBackingOff {
					logger.Logger.Info("Dumping diagnostics after backoff", "consecutiveFails", dm.consecutiveFails)
					dm.PluginMgr.DiagnosticDump(dm.Cfg.DiagnosticDir)
					dm.backOffDuration = min(dm.backOffDuration*2, dm.maxBackOff)
					isBackingOff = false
					backoffTimer = nil
				} // isBackingOff
			} // select
		} // for
	}()

	go dm.StartTicker(ctx)
}
