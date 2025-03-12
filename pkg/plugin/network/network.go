package network

import (
	"bytes"
	"fmt"
	"io"
	"os"
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
			"check_network":   "Check Network Status",
			"diagnostic_dump": "Dump network status to a directory",
		},
		cfg: config,
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

func (np *NetworkPlugin) Execute(command string, args ...string) (string, error) {
	switch command {
	case "check_network":
		output, err := np.run_ss("")
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "diagnostic_dump":
		if len(args) < 1 {
			return "", fmt.Errorf("no diagnostic directory provided")
		}
		network_dir := fmt.Sprintf("%s/network", args[0])
		err := os.MkdirAll(network_dir, 0755)
		if err != nil {
			return "", err
		}

		output, err := np.run_ss("")
		if err != nil {
			return "", err
		}

		err = os.WriteFile(fmt.Sprintf("%s/ss.txt", network_dir), output, 0644)
		if err != nil {
			return "", err
		}
		return "OK", nil
	default:
		return "", fmt.Errorf("command not implemented")
	}
}
