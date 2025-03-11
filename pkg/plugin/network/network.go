package network

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"

	"gitlab.cern.ch/eos/argeos/internal/logger"
	"gitlab.cern.ch/eos/argeos/pkg/plugin"
)

type NetworkPlugin struct {
	name        string
	commandHelp map[string]string
}

func NewPlugin() plugin.Plugin {
	return &NetworkPlugin{
		name: "network",
		commandHelp: map[string]string{
			"check network": "Check Network Status",
		},
	}
}

func (np *NetworkPlugin) Name() string {
	return "Linux"
}

func (np *NetworkPlugin) run_ss(args string) ([]byte, error) {
	if args == "" {
		args = "-tunap"
	}

	logger.Logger.Debug("Running ss with args", "args", args)
	cmd := exec.Command("ss", args)
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.Logger.Error("Error running ss", "error", err)
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		logger.Logger.Error("Error starting ss command", "error", err)
		return nil, err
	}

	out, err := io.ReadAll(outPipe)
	if err != nil {
		logger.Logger.Error("Error reading output from ss", "error", err)
		return nil, err
	}

	err = cmd.Wait()
	if err != nil {
		logger.Logger.Error("Error waiting for ss command", "error", err)
		return nil, err
	}

	return bytes.TrimSpace(out), nil

}

func (np *NetworkPlugin) HealthCheck() plugin.HealthStatus {
	logger.Logger.Debug("Running Network plugin")

	_, err := np.run_ss("")
	if err != nil {
		return plugin.HealthERROR(err.Error())
	}
	return plugin.HealthOK("OK")
}

func (np *NetworkPlugin) CommandHelp() map[string]string {
	return np.commandHelp
}

func (np *NetworkPlugin) Execute(command string, args ...string) string {
	switch command {
	case "check_network":
		output, err := np.run_ss("")
		if err != nil {
			return fmt.Sprintf("Error running ss: %s", err)
		}
		return string(output)
	default:
		return "Not Implemented!"
	}
}
