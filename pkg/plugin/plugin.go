package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gitlab.cern.ch/eos/argeos/internal/logger"
)

type HealthState int

const (
	StateOK HealthState = iota
	StateWARN
	StateERROR
)

type HealthStatus struct {
	State  HealthState `json:"state"`
	Detail string      `json:"detail"`
}

func HealthOK(status string) HealthStatus {
	return HealthStatus{State: StateOK, Detail: status}
}

func HealthWARN(status string) HealthStatus {
	return HealthStatus{State: StateWARN, Detail: status}
}

func HealthERROR(status string) HealthStatus {
	return HealthStatus{State: StateERROR, Detail: status}
}

type Plugin interface {
	Name() string // Name of the plugin
	HealthCheck() HealthStatus
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
	plugins []Plugin
}

func NewManager() *PluginManager {
	return &PluginManager{}
}

func (pm *PluginManager) Register(plugin Plugin) {
	pm.plugins = append(pm.plugins, plugin)
}

func (pm *PluginManager) ExecuteCommand(command string, args ...string) string {

	var result string

	for _, plugin := range pm.plugins {

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

	return result
}

func (pm *PluginManager) SupportedCommands() string {
	commands := make([]string, 0)
	for _, plugin := range pm.plugins {
		commands = append(commands, SupportedCommands(plugin)...)
	}
	bytes, err := json.Marshal(commands)
	if err != nil {
		return "Error encoding supported commands"
	}
	return string(bytes)
}

func (pm *PluginManager) HealthCheck() []HealthStatus {

	var result []HealthStatus
	for _, plugin := range pm.plugins {
		result = append(result, plugin.HealthCheck())
	}
	return result
}

func (pm *PluginManager) DiagnosticDump() string {
	// TODO: make this configurable
	dump_base_dir := "/tmp/eos-diagnostics"
	dump_dir_name := fmt.Sprintf("%s/dump-%s", dump_base_dir, time.Now().Format("20060102T150405"))
	err := os.MkdirAll(dump_dir_name, 0755)
	if err != nil {
		logger.Logger.Error("Error creating dump directory", "error", err)
	}
	return pm.ExecuteCommand("diagnostic_dump", dump_dir_name)
}
