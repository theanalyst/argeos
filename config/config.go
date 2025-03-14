package config

import (
	"encoding/json"
	"os"

	"gitlab.cern.ch/eos/argeos/internal/logger"
	list "gitlab.cern.ch/eos/argeos/internal/utils"
)

type ServerConfig struct {
	Address     string `json:"host"`
	AdminSocket string `json:"admin_socket"`
}

type NatsConfig struct {
	Servers list.StringList `json:"servers"`
}

type Config struct {
	Server  ServerConfig              `json:"server"`
	Nats    NatsConfig                `json:"nats"`
	Plugins map[string]map[string]any `json:"plugins"`
}

var defaultConfig Config = Config{
	Server: ServerConfig{
		Address:     ":9999",
		AdminSocket: "/var/run/argeos.asok",
	},
}

func overrideDefaults(config *Config) {
	if config.Server.Address == "" {
		config.Server.Address = defaultConfig.Server.Address
	}
}

func Configure(jsonString []byte) Config {
	var config Config
	err := json.Unmarshal([]byte(jsonString), &config)

	if err != nil {
		logger.Logger.Error("Parsing JSON config", "error", err)
		return defaultConfig
	}
	overrideDefaults(&config)
	return config
}

func ConfigurefromFile(filename string) Config {
	file, err := os.ReadFile(filename)
	if err != nil {
		logger.Logger.Warn("Reading file - using defaults", "error", err)
		return defaultConfig
	}
	return Configure(file)
}
