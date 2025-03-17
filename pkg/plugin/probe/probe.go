package probe

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"gitlab.cern.ch/eos/argeos/config"
	"gitlab.cern.ch/eos/argeos/internal/common"
	"gitlab.cern.ch/eos/argeos/internal/logger"
	"gitlab.cern.ch/eos/argeos/pkg/plugin"
	"gitlab.cern.ch/eos/ops/probe"
)

type ProbePlugin struct {
	name        string
	commandHelp map[string]string
	cfg         config.Config
}

func NewPlugin(config config.Config) plugin.Plugin {
	return &ProbePlugin{
		name: "probe",
		commandHelp: map[string]string{
			"check_probe": "Check Probe Status",
		},
		cfg: config,
	}
}

func (p *ProbePlugin) Name() string {
	return "eosmon"
}

func (p *ProbePlugin) HealthCheck() common.HealthStatus {
	logger.Logger.Debug("Running Probe plugin")

	store, _ := probe.NewStore(p.cfg.Nats.Servers)
	hostname, _ := os.Hostname() // can be any MGM hostname like: eosalice-ns-ip700, eosatlas-ns-ip700

	return p.GetManualUpdates(store, hostname)
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

// TODO: Make Probe a standalone component instead of a plugin
func (p *ProbePlugin) StartProbe() {
	store, _ := probe.NewStore(p.cfg.Nats.Servers)
	hostname, _ := os.Hostname() // can be any MGM hostname like: eosalice-ns-ip700, eosatlas-ns-ip700
	go func() {
		for {
			p.GetAutomaticUpdates(store, hostname)
			time.Sleep(5 * time.Second)
		}
	}()
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
		if strings.Contains(hostname, target) {
			info, err := store.GetProbeInfo(target)
			if err != nil {
				logger.Logger.Error("Error running healthcheck", "error", err)
				return common.HealthERROR(err.Error())
			}
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

func (p *ProbePlugin) PollforUpdates(store *probe.Store, hostname string) common.HealthStatus {
	if store == nil {
		logger.Logger.Error("No Probe store, not running Probe")
		return common.HealthERROR("No Probe store")
	}

	logger.Logger.Info("Polling for updates")
	targets, err := store.ListTargets()
	if err != nil {
		return common.HealthERROR(err.Error())
	}

	for _, target := range targets {
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

func (p *ProbePlugin) Execute(command string, args ...string) (string, error) {
	switch command {
	case "check_probe":
		return p.HealthCheck().Detail, nil
	default:
		return "", fmt.Errorf("command not implemented")
	}
}
