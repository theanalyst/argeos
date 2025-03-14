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

func HealthStateString(state HealthState) string {
	switch state {
	case StateOK:
		return "OK"
	case StateWARN:
		return "WARN"
	case StateERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type HealthStatus struct {
	State       HealthState `json:"state"`
	StateString string      `json:"state_string"`
	Name        string      `json:"plugin_name"`
	Detail      string      `json:"detail"`
}

// TODO: Add a helper struct that carries the plugin name
// and make these functions easier
func HealthOK(status string) HealthStatus {
	return HealthStatus{State: StateOK, StateString: HealthStateString(StateOK), Detail: status}
}

func HealthWARN(status string) HealthStatus {
	return HealthStatus{State: StateWARN, StateString: HealthStateString(StateWARN), Detail: status}
}

func HealthERROR(status string) HealthStatus {
	return HealthStatus{State: StateERROR, StateString: HealthStateString(StateERROR), Detail: status}
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
	if result == "" {
		return fmt.Sprintf("Command %s not supported", command)
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
	logger.Logger.Info("Running healthcheck")
	for _, plugin := range pm.plugins {
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
