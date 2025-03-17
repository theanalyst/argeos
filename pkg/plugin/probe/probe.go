package probe

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gitlab.cern.ch/eos/argeos/config"
	"gitlab.cern.ch/eos/argeos/internal/common"
	"gitlab.cern.ch/eos/argeos/internal/logger"
	"gitlab.cern.ch/eos/argeos/pkg/plugin"
	"gitlab.cern.ch/eos/ops/probe"
)

type ProbePlugin struct {
	name        string
	commandHelp map[string]string
	nats_cfg    config.NatsConfig
}

func NewPlugin(config config.Config) plugin.Plugin {
	_nats_cfg := config.Nats
	if _nats_cfg.Target == "" {
		_nats_cfg.Target, _ = os.Hostname()
	}
	return &ProbePlugin{
		name: "probe",
		commandHelp: map[string]string{
			"check_probe": "Check Probe Status",
		},
		nats_cfg: _nats_cfg,
	}
}

func (p *ProbePlugin) Name() string {
	return p.name
}

func (p *ProbePlugin) HealthCheck() common.HealthStatus {
	logger.Logger.Debug("Running Probe plugin")

	store, _ := probe.NewStore(p.nats_cfg.Servers)
	hostname, _ := os.Hostname() // can be any MGM hostname like: eosalice-ns-ip700, eosatlas-ns-ip700

	return p.GetManualUpdates(store, hostname).WithComponent(p.Name())
}

func (p *ProbePlugin) CommandHelp() map[string]string {
	return p.commandHelp
}

func (p *ProbePlugin) GetAutomaticUpdates(store *probe.Store, hostname string) common.HealthStatus {
	if store == nil {
		logger.Logger.Error("No Probe store, not running Probe")
		return common.HealthERROR("No Probe store")
	}

	logger.Logger.Info("Running Probe plugin")
	lis, err := store.Listener(probe.WithName("argeos"))
	if err != nil {
		logger.Logger.Error("Error creating listener", "error", err)
		return common.HealthERROR(err.Error())
	}

	for _target := range lis.Updates() {
		target := _target.Target
		if strings.Contains(hostname, target) {
			info, err := store.GetProbeInfo(target)
			if err != nil {
				logger.Logger.Error("Error running healthcheck", "error", err)
				continue
			}
			logger.Logger.Debug("Checking probe info automatically", "info", info)
			if info.Available {
				logger.Logger.Info(target + " is working")
			} else {
				logger.Logger.Warn(target + " is not working")
				cmd := exec.Command("ping", "-c", "4", target)
				// Get the command output
				output, err := cmd.CombinedOutput()
				if err != nil {
					logger.Logger.Error("Error running ping", "error", err)

				}
				logger.Logger.Warn(string(output))
			}
		}
	}
	return common.HealthOK("OK")
}

func (p *ProbePlugin) isTarget(target string) bool {
	return strings.Contains(p.nats_cfg.Target, target)
}

func probeHealthStatus(info *probe.ProbeInfo) common.HealthStatus {
	if info.Available {
		status, err := info.AvailabilityInfo()
		if err != nil {
			return common.HealthWARN(err.Error())
		}

		return common.HealthOK(status)
	}
	status, err := info.ErrorDescription()
	if err != nil {
		logger.Logger.Error("Error getting availability status", "error", err)
		return common.HealthFAIL(err.Error())
	}
	return common.HealthFAIL(status)
}

func (p *ProbePlugin) Start(ctx context.Context, updateChannel chan<- common.HealthStatus) error {
	store, err := probe.NewStore(p.nats_cfg.Servers)
	if err != nil {
		logger.Logger.Error("Error creating store", "error", err)
		return err
	}

	logger.Logger.Info("Starting Probe diagnostics plugin")
	listener, err := store.Listener(probe.WithName("argeos"))
	if err != nil {
		logger.Logger.Error("Error creating listener", "error", err)
		return err
	}

	go func() {
		defer listener.Close()
		for {
			select {
			case <-ctx.Done():
				logger.Logger.Info("Stopping Probe plugin")
				return
			case _target := <-listener.Updates():
				target := _target.Target

				if p.isTarget(target) {
					info, err := store.GetProbeInfo(target)
					if err != nil {
						logger.Logger.Error("Error running healthcheck", "error", err)
						continue
					}
					updateChannel <- probeHealthStatus(info)
					logger.Logger.Debug("Probe status", "status", info)
				}
			}
		}
	}()

	<-ctx.Done()
	logger.Logger.Info("Probe plugin stopped")
	return nil

}

func (p *ProbePlugin) Stop() {
	logger.Logger.Info("Stopping Probe diagnostics plugin")
}

func (p *ProbePlugin) GetManualUpdates(store *probe.Store, hostname string) common.HealthStatus {
	if store == nil {
		return common.HealthERROR("No Probe store")
	}

	targets, err := store.ListTargets()
	if err != nil {
		return common.HealthERROR(err.Error())
	}

	for _, target := range targets {
		if p.isTarget(target) {
			info, err := store.GetProbeInfo(target)
			if err != nil {
				logger.Logger.Error("Error running healthcheck", "error", err)
				return common.HealthERROR(err.Error())
			}

			healthStatus := probeHealthStatus(info)
			logger.Logger.Debug("Got health status", "status", healthStatus)
			return healthStatus
		}
	}
	return common.HealthOK("OK")
}

func (p *ProbePlugin) Execute(command string, args ...string) (string, error) {
	switch command {
	case "check_probe":
		return p.HealthCheck().Detail, nil
	default:
		return "", fmt.Errorf("command not implemented")
	}
}
