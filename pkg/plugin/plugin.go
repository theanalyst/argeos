package plugin

type HealthState int

const (
	StateOK HealthState = iota
	StateWARN
	StateERROR
)

type HealthStatus struct {
	State  HealthState  `json:"state"`
	Detail string       `json:"detail"`
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
	Execute(command string, args ...string) string
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
				result += plugin.Execute(command, args...)
				result +="\n"
			}
		}
	}

	return result
}

func (pm *PluginManager) HealthCheck() []HealthStatus {

	var result []HealthStatus
	for _, plugin := range pm.plugins {
		result  = append(result, plugin.HealthCheck())
	}
	return result
}
