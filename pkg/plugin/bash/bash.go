package bash

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gitlab.cern.ch/eos/argeos/config"
	"gitlab.cern.ch/eos/argeos/internal/common"
	"gitlab.cern.ch/eos/argeos/internal/logger"
	"gitlab.cern.ch/eos/argeos/pkg/plugin"
)

type PluginConfig struct {
	ScriptDir string            `json:"script_dir"`
	EnvVars   map[string]string `json:"env_vars"`
}

const DefaultScriptDir = "/usr/share/argeos/scripts"

type BashPlugin struct {
	name        string
	commandHelp map[string]string
	config      PluginConfig
}

func extractConfig(cfg config.Config) PluginConfig {
	pluginConfig, exists := cfg.Plugins["bash"]
	if !exists {
		return PluginConfig{ScriptDir: DefaultScriptDir}
	}
	cfgBytes, err := json.Marshal(pluginConfig)
	if err != nil {
		logger.Logger.Error("Error marshalling plugin config", "error", err)
		return PluginConfig{ScriptDir: DefaultScriptDir}
	}
	var config PluginConfig
	err = json.Unmarshal(cfgBytes, &config)
	if err != nil {
		logger.Logger.Error("Error unmarshalling plugin config", "error", err)
		return PluginConfig{ScriptDir: DefaultScriptDir}
	}
	return config
}

func NewPlugin(cfg config.Config) plugin.Plugin {
	bash_cfg := extractConfig(cfg)
	return &BashPlugin{
		name: "bash",
		commandHelp: map[string]string{
			"run_script":      "Run a bash script",
			"diagnostic_dump": "Run all diagnostic scripts",
		},
		config: bash_cfg,
	}
}

func (bp *BashPlugin) Name() string {
	return bp.name
}

func (bp *BashPlugin) CommandHelp() map[string]string {
	return bp.commandHelp
}

func (bp *BashPlugin) getScripts() ([]string, error) {
	files, err := os.ReadDir(bp.config.ScriptDir)
	if err != nil {
		logger.Logger.Error("Error reading script directory", "error", err)
		return nil, err
	}

	scripts := make([]string, 0, len(files))
	for _, file := range files {

		if file.IsDir() {
			logger.Logger.Debug("Skipping directory", "directory", file.Name())
			continue
		}

		fileinfo, err := file.Info()
		if err != nil {
			logger.Logger.Error("Error getting file info", "file", file.Name(), "error", err)
			continue
		}
		if fileinfo.Mode()&0111 == 0 {
			logger.Logger.Info("Skipping non-executable file", "file", file.Name(), "permissions", fileinfo.Mode().String())
			continue
		}
		scripts = append(scripts, file.Name())
	}
	sort.Strings(scripts) // sort lexically like unix tools
	return scripts, nil
}

func (bp *BashPlugin) runScripts(script_env []string) (string, error) {
	files, err := bp.getScripts()
	if err != nil || len(files) == 0 {
		return "", err
	}

	var output strings.Builder

	logger.Logger.Debug("Running scripts", "scripts", files)
	for _, file := range files {
		cmd := exec.Command(filepath.Join(bp.config.ScriptDir, file))
		cmd.Env = append(os.Environ(), script_env...)
		out, err := cmd.CombinedOutput()

		if err != nil {
			logger.Logger.Error("Error running script", "script", file, "error", err)
			output.WriteString(fmt.Sprintf("Error running script %s: %s\n", file, err))
		}
		output.WriteString(fmt.Sprintf("=== Running %s ===\n%s\n", file, string(out)))
	}
	return output.String(), nil
}

func (bp *BashPlugin) Execute(command string, args ...string) (string, error) {
	switch command {
	case "run_script":
		fallthrough
	case "diagnostic_dump":
		if len(args) < 1 {
			return "", fmt.Errorf("no diagnostic directory provided")
		}
		scriptEnv := []string{
			fmt.Sprintf("DUMP_DIR=%s", args[0]),
		}
		return bp.runScripts(scriptEnv)
	default:
		return "", fmt.Errorf("command not implemented")
	}
}

func (bp *BashPlugin) HealthCheck() common.HealthStatus {
	if _, err := os.Stat(bp.config.ScriptDir); os.IsNotExist(err) {
		return common.HealthERROR("Script directory does not exist")
	}

	filelist, err := bp.getScripts()

	if err != nil {
		return common.HealthERROR("Error reading script directory")
	}
	if len(filelist) == 0 {
		return common.HealthWARN("No scripts found in script directory")
	}
	return common.HealthOK("Bash plugin is healthy")
}
