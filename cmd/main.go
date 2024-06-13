package main

import (
	"flag"
	"gitlab.cern.ch/eos/argeos/internal/config"
	"gitlab.cern.ch/eos/argeos/internal/server"
)

func main() {
	var configpath string
	flag.StringVar(&configpath, "c", "/etc/argeos.config.json",
		"Path to config file [/etc/argeos.config.json]")

	config := config.ConfigurefromFile("config.json")
	server.StartServer(config.Server.Address)
}
