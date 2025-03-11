package network

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"

	"gitlab.cern.ch/eos/argeos/config"
	"gitlab.cern.ch/eos/argeos/internal/logger"
	"gitlab.cern.ch/eos/argeos/pkg/plugin"
)

type NetworkPlugin struct {
	name        string
	commandHelp map[string]string
	cfg         config.Config
}

func NewPlugin(config config.Config) plugin.Plugin {
	return &NetworkPlugin{
		name: "network",
		commandHelp: map[string]string{
			"check network": "Check Network Status",
		},
		cfg: config,
	}
}

func (np *NetworkPlugin) Name() string {
	return "Linux"
}

func (np *NetworkPlugin) HealthCheck() plugin.HealthStatus {
	logger.Logger.Debug("Running Network plugin")

	cmd := exec.Command("ss", "-tunap")
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.Logger.Error("Error running ss", "error", err)
		return plugin.HealthERROR(err.Error())
	}

	err = cmd.Start()
	if err != nil {
		logger.Logger.Error("Error starting ss command", "error", err)
		return plugin.HealthERROR(err.Error())
	}

	out, err := io.ReadAll(outPipe)
	if err != nil {
		logger.Logger.Error("Error reading output from ss", "error", err)
		return plugin.HealthERROR(err.Error())
	}
	var formattedOutput bytes.Buffer
	_, err = fmt.Fprintf(&formattedOutput, "%s\n", string(bytes.TrimSpace(out)))

	err = cmd.Wait()
	if err != nil {
		logger.Logger.Error("Error waiting for ss command", "error", err)
		return plugin.HealthERROR(err.Error())
	}

	return plugin.HealthOK(formattedOutput.String())
}

func (np *NetworkPlugin) CommandHelp() map[string]string {
	return np.commandHelp
}

func (np *NetworkPlugin) Execute(command string, args ...string) string {
	switch command {
	case "check_network":
		output, err := exec.Command("ss", "-tunap").Output()
		if err != nil {
			logger.Logger.Error("Error running ss", "error", err)

		}
		return string(output)
	default:
		return "Not Implemented!"
	}
}
