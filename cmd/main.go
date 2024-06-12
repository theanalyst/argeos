package main

import (
	"gitlab.cern.ch/eos/argeos/internal/config"
	"gitlab.cern.ch/eos/argeos/internal/server"
)

func main() {

	config := config.ConfigurefromFile("config.json")
	server.StartServer(config.Server.Address)
}
