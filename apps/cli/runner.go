package main

import (
	"github.com/oni1997/gatewayx/internal/config"
	"github.com/oni1997/gatewayx/internal/server"
)

func loadConfig(configFile string) (*config.Config, error) {
	return config.LoadConfig(configFile)
}

func runServer(configFile string) {
	server.Run(configFile, server.Version{
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
	})
}
