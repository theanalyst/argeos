package config

import (
	"encoding/json"
	"os"

	"gitlab.cern.ch/eos/argeos/internal/logger"
	list "gitlab.cern.ch/eos/argeos/internal/utils"
)

var CmdLogFile string

type ServerConfig struct {
	Address            string `json:"host"`
	AdminSocket        string `json:"admin_socket"`
	DiagnosticDir      string `json:"diagnostic_dir"`
	DiagnosticInterval int32  `json:"diagnostic_interval"`
	LogLevel           string `json:"log_level"`
	LogFile            string `json:"log_file"`
}

type NatsConfig struct {
	Servers list.StringList `json:"servers"`
	Target  string          `json:"target"`
}

type Config struct {
	Server  ServerConfig              `json:"server"`
	Nats    NatsConfig                `json:"nats"`
	Plugins map[string]map[string]any `json:"plugins"`
}

var defaultConfig Config = Config{
	Server: ServerConfig{
		Address:            ":9999",
		AdminSocket:        "/var/run/argeos.asok",
		DiagnosticDir:      "/var/lib/argeos/diagnostics",
		DiagnosticInterval: 300,
		LogFile:            "/var/log/argeos/argeos.log",
	},
}

func overrideDefaults(config *Config) {
	if config.Server.Address == "" {
		config.Server.Address = defaultConfig.Server.Address
	}
	if config.Server.AdminSocket == "" {
		config.Server.AdminSocket = defaultConfig.Server.AdminSocket
	}
	if config.Server.DiagnosticDir == "" {
		config.Server.DiagnosticDir = defaultConfig.Server.DiagnosticDir
	}
	if config.Server.DiagnosticInterval == 0 {
		config.Server.DiagnosticInterval = defaultConfig.Server.DiagnosticInterval
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
	if CmdLogFile == "" && config.Server.LogFile != "" {
		logger.Init(config.Server.LogFile)
	}
	if config.Server.LogLevel != "" {
		logger.SetLogLevelfromString(config.Server.LogLevel)
	}
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
