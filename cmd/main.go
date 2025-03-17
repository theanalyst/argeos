package main

import (
	"flag"

	"gitlab.cern.ch/eos/argeos/config"
	"gitlab.cern.ch/eos/argeos/internal/logger"
	"gitlab.cern.ch/eos/argeos/internal/server"
	"gitlab.cern.ch/eos/argeos/pkg/plugin"
	"gitlab.cern.ch/eos/argeos/pkg/plugin/bash"
	"gitlab.cern.ch/eos/argeos/pkg/plugin/network"
	"gitlab.cern.ch/eos/argeos/pkg/plugin/probe"
)

func main() {

	var configpath string
	var logfile string
	flag.StringVar(&configpath, "c", "/etc/argeos.config.json",
		"Path to config file [/etc/argeos.config.json]")
	flag.StringVar(&logfile, "logfile", "", "Path to log file")
	flag.Parse()

	logger.Init(logfile)
	config := config.ConfigurefromFile(configpath)

	pluginmgr := plugin.NewManager()
	networkplugin := network.NewPlugin(config)
	pluginmgr.Register(networkplugin)

	probeplugin := probe.NewPlugin(config)
	pluginmgr.Register(probeplugin)

	bashplugin := bash.NewPlugin(config)
	pluginmgr.Register(bashplugin)

	server := server.NewServer(config.Server, pluginmgr)
	server.Start()
}
