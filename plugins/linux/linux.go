package linux

import (
	"os/exec"

	"gitlab.cern.ch/eos/argeos/internal/logger"
	"gitlab.cern.ch/eos/argeos/pkg/plugin"
)

type LinuxPlugin struct {
	path string
}

func New(path string) plugin.Plugin {
	return &LinuxPlugin{path: path}
}

func (l *LinuxPlugin) Name() string {
	return "Linux"
}


func (l *LinuxPlugin) HealthCheck() plugin.HealthStatus {
	logger.Logger.Info("Running Network plugin")

	output, err := exec.Command("ss", "-tunap").Output()

	if err != nil {
		logger.Logger.Error("Error running ss", "error", err)
		return plugin.HealthERROR(err)
	}
	return plugin.HealthOK(output)
}

func (l *LinuxPlugin) CommandHelp() map[string]string {
	m := make(map[string]string)
	m["check network"] = "Check Network status"
	return m
}

func (l *LinuxPlugin) Execute(command string, args ...string) string {
	switch (command) {
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
