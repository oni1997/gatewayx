package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a gatewayx.yaml configuration interactively",
	Long:  `Walk through a series of questions to generate a working gatewayx.yaml configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			output = "gatewayx.yaml"
		}

		cfg := interactiveConfig()
		if err := writeConfig(cfg, output); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nConfiguration written to %s\n", output)
		fmt.Printf("Run with: gatewayx serve -c %s\n", output)
	},
}

type route struct {
	Name         string            `yaml:"name"`
	ListenPath   string            `yaml:"listen_path"`
	UpstreamURLs []string          `yaml:"upstream_urls"`
	Methods      []string          `yaml:"methods,omitempty"`
	StripPath    bool              `yaml:"strip_path,omitempty"`
	Headers      map[string]string `yaml:"headers,omitempty"`
	Timeout      string            `yaml:"timeout,omitempty"`
	Auth         *authConfig       `yaml:"authentication,omitempty"`
	RateLimit    *rateLimit        `yaml:"rate_limit,omitempty"`
	Cache        *cacheConfig      `yaml:"cache,omitempty"`
	Websocket    bool              `yaml:"websocket,omitempty"`
}

type authConfig struct {
	Type    string            `yaml:"type"`
	Options map[string]string `yaml:"options,omitempty"`
}

type rateLimit struct {
	Rate   float64 `yaml:"rate"`
	Burst  int     `yaml:"burst"`
	PerIP  bool    `yaml:"per_ip,omitempty"`
	PerUser bool   `yaml:"per_user,omitempty"`
}

type cacheConfig struct {
	TTL string `yaml:"ttl"`
}

type initConfig struct {
	Server  serverConfig `yaml:"server"`
	Routes  []route      `yaml:"routes"`
	Logging loggingConfig `yaml:"logging"`
	Metrics metricsConfig `yaml:"metrics"`
	Health  healthConfig  `yaml:"health"`
}

type serverConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type loggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type metricsConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

type healthConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

func interactiveConfig() *initConfig {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== GatewayX Configuration Generator ===")

	port := askInt(reader, "Gateway port", 8080)

	var routes []route
	for {
		fmt.Println("\n--- Route ---")
		name := askString(reader, "Route name", fmt.Sprintf("route-%d", len(routes)+1))
		listenPath := askString(reader, "Listen path", "/")
		upstream := askString(reader, "Upstream URL (comma-separated for multiple)", "http://localhost:3000")

		upstreams := splitAndTrim(upstream)

		authType := askString(reader, "Auth type (none/jwt/api_key/basic)", "none")
		var auth *authConfig
		switch strings.ToLower(authType) {
		case "jwt":
			secret := askString(reader, "JWT secret", "change-me")
			auth = &authConfig{Type: "jwt", Options: map[string]string{"secret": secret, "algorithm": "HS256"}}
		case "api_key":
			header := askString(reader, "API key header", "X-API-Key")
			key := askString(reader, "Valid API key", "sk-test-123")
			auth = &authConfig{Type: "api_key", Options: map[string]string{"header": header, key: "owner"}}
		case "basic":
			user := askString(reader, "Username", "admin")
			pass := askString(reader, "Password", "password")
			auth = &authConfig{Type: "basic", Options: map[string]string{user: pass}}
		}

		rateStr := askString(reader, "Rate limit (req/s, 0 for none)", "0")
		var rl *rateLimit
		if r := parseFloat(rateStr); r > 0 {
			burst := askInt(reader, "Burst", int(r)*2)
			perIP := askBool(reader, "Per-IP rate limit?", false)
			rl = &rateLimit{Rate: r, Burst: burst, PerIP: perIP}
		}

		cacheStr := askString(reader, "Cache TTL (e.g. 30s, empty for none)", "")
		var cache *cacheConfig
		if cacheStr != "" {
			cache = &cacheConfig{TTL: cacheStr}
		}

		ws := askBool(reader, "WebSocket route?", false)

		routes = append(routes, route{
			Name:         name,
			ListenPath:   listenPath,
			UpstreamURLs: upstreams,
			Methods:      []string{"GET", "POST", "PUT", "DELETE"},
			StripPath:    strings.HasPrefix(listenPath, "/") && listenPath != "/",
			Timeout:      "30s",
			Auth:         auth,
			RateLimit:    rl,
			Cache:        cache,
			Websocket:    ws,
		})

		if !askBool(reader, "Add another route?", false) {
			break
		}
	}

	cfg := &initConfig{
		Server: serverConfig{Host: "0.0.0.0", Port: port},
		Routes: routes,
		Logging: loggingConfig{Level: "info", Format: "json"},
		Metrics: metricsConfig{Enabled: true, Port: 9090},
		Health:  healthConfig{Enabled: true, Path: "/health"},
	}

	return cfg
}

func writeConfig(cfg *initConfig, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func askString(r *bufio.Reader, prompt, def string) string {
	fmt.Printf("%s [%s]: ", prompt, def)
	line, _ := r.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}

func askInt(r *bufio.Reader, prompt string, def int) int {
	fmt.Printf("%s [%d]: ", prompt, def)
	line, _ := r.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	var n int
	if _, err := fmt.Sscanf(line, "%d", &n); err != nil {
		return def
	}
	return n
}

func askBool(r *bufio.Reader, prompt string, def bool) bool {
	defStr := "n"
	if def {
		defStr = "y"
	}
	fmt.Printf("%s (y/n) [%s]: ", prompt, defStr)
	line, _ := r.ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))
	if line == "" {
		return def
	}
	return line == "y" || line == "yes"
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}

func parseFloat(s string) float64 {
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		return 0
	}
	return f
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("output", "o", "", "Output file path")
}
