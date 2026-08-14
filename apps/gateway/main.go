package main

import (
	"github.com/oni1997/gatewayx/internal/server"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	server.Run("", server.Version{
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
	})
}
