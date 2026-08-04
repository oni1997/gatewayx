package main

import (
	"fmt"
	"os"

	"github.com/oni1997/gatewayx/internal/config"
)

func loadConfig(configFile string) (*config.Config, error) {
	return config.LoadConfig(configFile)
}

func runServer() {
	fmt.Fprintln(os.Stderr, "Running gateway server...")
	fmt.Fprintln(os.Stderr, "Build the gateway binary with: go build -o bin/gatewayx ./apps/gateway")
	os.Exit(1)
}
