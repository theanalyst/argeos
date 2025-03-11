package probe

import (
	"os"
	"os/exec"
	"strings"

	"gitlab.cern.ch/eos/argeos/config"
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
			"check probe": "Check Probe Status",
		},
		cfg: config,
	}
}

func (p *ProbePlugin) Name() string {
	return "eosmon"
}

func (p *ProbePlugin) HealthCheck() plugin.HealthStatus {
	logger.Logger.Debug("Running Probe plugin")

	store, _ := probe.NewStore(p.cfg.Nats.Servers)
	hostname, _ := os.Hostname() // can be any MGM hostname like: eosalice-ns-ip700, eosatlas-ns-ip700

	return p.GetManualUpdates(store, hostname)
}

func (p *ProbePlugin) CommandHelp() map[string]string {
	return p.commandHelp
}

func (p *ProbePlugin) Execute(command string, args ...string) string {
	switch command {
	case "check_probe":
		output, err := exec.Command("ss", "-tunap").Output()
		if err != nil {
			logger.Logger.Error("Error running ss", "error", err)

		}
		return string(output)
	default:
		return "Not Implemented!"
	}
}

func (p *ProbePlugin) GetAutomaticUpdates(store *probe.Store, hostname string) plugin.HealthStatus {
	lis, err := store.Listener(probe.WithName("argeos"))
	if err != nil {
		return plugin.HealthERROR(err.Error())
	}

	for target := range lis.Updates() {
		if strings.Contains(hostname, target.Target) {
			info, err := store.GetProbeInfo(hostname)
			if err != nil {
				logger.Logger.Error("Error running healthcheck", "error", err)
			}
			if info.Available {
				logger.Logger.Info(target.Target + " is working")
			} else {
				logger.Logger.Warn(target.Target + " is not working")
				cmd := exec.Command("ping", "-c", "4", target.Target)
				// Get the command output
				output, err := cmd.CombinedOutput()
				if err != nil {
					logger.Logger.Error("Error running ping", "error", err)

				}
				logger.Logger.Warn(string(output))
			}
		}
	}
	return plugin.HealthOK("OK")
}

func (p *ProbePlugin) GetManualUpdates(store *probe.Store, hostname string) plugin.HealthStatus {
	targets, err := store.ListTargets()
	if err != nil {
		return plugin.HealthERROR(err.Error())
	}

	for _, target := range targets {
		if strings.Contains(hostname, target) {
			info, err := store.GetProbeInfo(target)
			if err != nil {
				logger.Logger.Error("Error running healthcheck", "error", err)
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
	return plugin.HealthOK("OK")
}
