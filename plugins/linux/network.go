package network

import (
	"os/exec"

	"gitlab.cern.ch/eos/argeos/internal/logger"
	"gitlab.cern.ch/eos/argeos/pkg/plugin"
)

type NetworkPlugin struct {
}

func New(path string) plugin.Plugin {
	return &NetworkPlugin{path: path}
}

func (np *NetworkPlugin) Name() string {
	return "Linux"
}

func (np *NetworkPlugin) HealthCheck() plugin.HealthStatus {
	logger.Logger.Info("Running Network plugin")

	output, err := exec.Command("ss", "-tunap").Output()

	if err != nil {
		logger.Logger.Error("Error running ss", "error", err)
		return plugin.HealthERROR(err)
	}
	return plugin.HealthOK(output)
}

func (np *NetworkPlugin) CommandHelp() map[string]string {
	m := make(map[string]string)
	m["check network"] = "Check Network status"
	return m
}

func (np *NetworkPlugin) Execute(command string, args ...string) string {
	switch command {
	case "check network":
		output, err := exec.Command("ss", "-tunap").Output()
		if err != nil {
			logger.Logger.Error("Error running ss", "error", err)

		}
		return output
	default:
		return "Not Implemented!"
	}
}
