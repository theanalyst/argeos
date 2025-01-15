package main

import (
	"flag"

	"gitlab.cern.ch/eos/argeos/internal/config"
	"gitlab.cern.ch/eos/argeos/internal/logger"
	"gitlab.cern.ch/eos/argeos/internal/server"
	"gitlab.cern.ch/eos/argeos/pkg/plugin"
	"gitlab.cern.ch/eos/argeos/pkg/plugin/network"
)

func main() {

	var configpath string
	var logfile string
	flag.StringVar(&configpath, "c", "/etc/argeos.config.json",
		"Path to config file [/etc/argeos.config.json]")
	flag.StringVar(&logfile, "logfile", "", "Path to log file")
	flag.Parse()


	pluginmgr := plugin.NewManager()
	networkplugin := network.NewPlugin()
	pluginmgr.Register(networkplugin)

	logger.Init(logfile)
	config := config.ConfigurefromFile(configpath)
	server := server.Server{config.Server}
	server.Start()
}
