package network

import (
	"os/exec"

	"gitlab.cern.ch/eos/argeos/internal/logger"
	"gitlab.cern.ch/eos/argeos/pkg/plugin"
)

type NetworkPlugin struct {
	name string
	commandHelp map[string]string
}

func NewPlugin() plugin.Plugin {
	return &NetworkPlugin{
		name:"network",
		commandHelp: map[string]string{
			"check network": "Check Network Status",
		},
	}
}

func (np *NetworkPlugin) Name() string {
	return "Linux"
}

func (np *NetworkPlugin) HealthCheck() plugin.HealthStatus {
	logger.Logger.Info("Running Network plugin")

	output, err := exec.Command("ss", "-tunap").Output()

	if err != nil {
		logger.Logger.Error("Error running ss", "error", err)
		return plugin.HealthERROR(err.Error())
	}
	return plugin.HealthOK(string(output))
}

func (np *NetworkPlugin) CommandHelp() map[string]string {
	return np.commandHelp
}

func (np *NetworkPlugin) Execute(command string, args ...string) string {
	switch command {
	case "check network":
		output, err := exec.Command("ss", "-tunap").Output()
		if err != nil {
			logger.Logger.Error("Error running ss", "error", err)

		}
		return string(output)
	default:
		return "Not Implemented!"
	}
}
