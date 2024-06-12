package main

import (
	"gitlab.cern.ch/eos/argeos/internal/server"
)

func main() {
	server.StartServer(":9999")
}
