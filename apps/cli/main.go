package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gatewayx",
	Short: "GatewayX - Developer Infrastructure Platform",
	Long: `GatewayX is a high-performance, extensible API gateway
built for developer infrastructure.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GatewayX %s\n", version)
		fmt.Printf("  commit:    %s\n", commit)
		fmt.Printf("  build date: %s\n", buildDate)
	},
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the GatewayX proxy server",
	Long:  `Start the GatewayX reverse proxy server with the specified configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		if configFile != "" {
			_ = os.Setenv("GATEWAYX_CONFIG", configFile)
		}
		runServer()
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a GatewayX configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		if configFile == "" {
			configFile = "gatewayx.yaml"
		}

		cfg, err := loadConfig(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Configuration invalid: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Configuration is valid (%d routes configured)\n", len(cfg.Routes))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(validateCmd)

	serveCmd.Flags().StringP("config", "c", "", "Path to configuration file")
	validateCmd.Flags().StringP("config", "c", "", "Path to configuration file")
}
