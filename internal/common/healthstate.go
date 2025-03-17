package common

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
