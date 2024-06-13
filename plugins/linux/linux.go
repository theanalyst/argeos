package linux

import (
	"os"
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

func (l *LinuxPlugin) Run() error {
	logger.Logger.Info("Running Linux plugin")

	output, err := exec.Command("ss", "-tunap").Output()

	if err != nil {
		logger.Logger.Error("Error running ss", "error", err)
		return err
	}

	// save output to file
	filename := l.path + "/ss_output.txt"
	err = os.WriteFile(filename, output, 0644)
	if err != nil {
		logger.Logger.Error("Error writing to file", "error", err)
		return err
	}
	return nil
}
