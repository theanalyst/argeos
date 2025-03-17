package common

import (
	"context"
)

type HealthState int

const (
	StateOK HealthState = iota
	StateWARN
	StateFAIL  // Use this for Failing server component
	StateERROR // Use for failure in plugin execution, or other errors but not for health failure
)

func HealthStateString(state HealthState) string {
	switch state {
	case StateOK:
		return "OK"
	case StateWARN:
		return "WARN"
	case StateERROR:
		return "ERROR"
	case StateFAIL:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

type HealthStatus struct {
	State       HealthState `json:"state"`
	StateString string      `json:"state_string"`
	Name        string      `json:"component"`
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

func HealthFAIL(status string) HealthStatus {
	return HealthStatus{State: StateFAIL, StateString: HealthStateString(StateFAIL), Detail: status}
}

func (status HealthStatus) WithComponent(name string) HealthStatus {
	status.Name = name
	return status
}

type HealthDaemon interface {
	Name() string
	Start(ctx context.Context, updateChannel chan<- HealthStatus) error
	HealthCheck() HealthStatus
	Stop()
}
