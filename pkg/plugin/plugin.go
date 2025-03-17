package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gitlab.cern.ch/eos/argeos/internal/common"
	"gitlab.cern.ch/eos/argeos/internal/logger"
)

type Plugin interface {
	Name() string // Name of the plugin
	HealthCheck() common.HealthStatus
	CommandHelp() map[string]string
	Execute(command string, args ...string) (string, error)
}

func SupportedCommands(p Plugin) []string {
	commands := make([]string, 0, len(p.CommandHelp()))
	for command := range p.CommandHelp() {
		commands = append(commands, command)
	}
	return commands
}

type PluginManager struct {
	Plugins []Plugin
}

func NewManager() *PluginManager {
	return &PluginManager{}
}

func (pm *PluginManager) Register(plugin Plugin) {
	pm.Plugins = append(pm.Plugins, plugin)
}

func (pm *PluginManager) ExecuteCommand(command string, args ...string) string {

	var result string

	for _, plugin := range pm.Plugins {

		for _, cmd := range SupportedCommands(plugin) {
			if cmd == command {
				plugin_result, err := plugin.Execute(command, args...)
				if err != nil {
					logger.Logger.Error("Error executing command", "plugin", plugin.Name(), "command", command, "error", err)
					continue
				}
				result += plugin_result
				result += "\n"
			}
		}
	}
	if result == "" {
		return fmt.Sprintf("Command %s not supported", command)
	}
	return result
}

func (pm *PluginManager) SupportedCommands() string {
	commands := make([]string, 0)
	for _, plugin := range pm.Plugins {
		commands = append(commands, SupportedCommands(plugin)...)
	}
	bytes, err := json.Marshal(commands)
	if err != nil {
		return "Error encoding supported commands"
	}
	return string(bytes)
}

func (pm *PluginManager) HealthCheck() []common.HealthStatus {

	var result []common.HealthStatus
	logger.Logger.Info("Running healthcheck")
	for _, plugin := range pm.Plugins {
		plugin_health := plugin.HealthCheck()
		plugin_health.Name = plugin.Name()
		result = append(result, plugin_health)
		logger.Logger.Debug("Healthcheck done for ", "plugin", plugin_health.Name, "state", plugin_health.StateString)
	}
	return result
}

func (pm *PluginManager) DiagnosticDump(dump_base_dir string) string {
	// TODO: make this configurable
	dump_dir_name := fmt.Sprintf("%s/dumps/dump-%s", dump_base_dir, time.Now().Format("20060102T150405"))
	err := os.MkdirAll(dump_dir_name, 0755)
	if err != nil {
		logger.Logger.Error("Error creating dump directory", "error", err)
	}
	return pm.ExecuteCommand("diagnostic_dump", dump_dir_name)
}
