package plugin

type HealthState int

const (
	StateOK HealthState = iota
	StateWARN
	StateERROR
)

type HealthStatus struct {
	State  HealthState
	Detail string
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
	SupportedCommands() []string
}

func (p Plugin) SupportedCommands() []string {
	commands := make([]string, 0, len(p.CommandHelp()))
	for command := range p.CommandHelp() {
		commands = append(commands, command)
	}
	return commands
}
