package probe

import (
	"os/exec"

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

	return p.GetAutomaticUpdates(store, "eosp2")
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
		if target.Target == hostname {
			info, err := store.GetProbeInfo(hostname)
			if err != nil {
				logger.Logger.Error("Error running healthcheck", "error", err)
			}
			if info.Available {
				logger.Logger.Info(hostname + " is working")
			} else {
				logger.Logger.Warn(hostname + " is not working")
				cmd := exec.Command("ping", "-c", "4", hostname)
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
		if target == hostname {
			info, err := store.GetProbeInfo(hostname)
			if err != nil {
				logger.Logger.Error("Error running healthcheck", "error", err)
			}
			if info.Available {
				logger.Logger.Info(hostname + " is working")
			} else {
				logger.Logger.Warn(hostname + " is not working")
				cmd := exec.Command("ping", "-c", "4", hostname)
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
