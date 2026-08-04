package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Routes   []RouteConfig  `mapstructure:"routes"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	Plugins  PluginsConfig  `mapstructure:"plugins"`
	TLS      TLSConfig      `mapstructure:"tls"`
	Health   HealthConfig   `mapstructure:"health"`
	Security SecurityConfig `mapstructure:"security"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderBytes  int           `mapstructure:"max_header_bytes"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type RouteConfig struct {
	Name          string            `mapstructure:"name"`
	ListenPath    string            `mapstructure:"listen_path"`
	UpstreamURLs  []string          `mapstructure:"upstream_urls"`
	Methods       []string          `mapstructure:"methods"`
	Hosts         []string          `mapstructure:"hosts"`
	StripPath     bool              `mapstructure:"strip_path"`
	PreserveHost  bool              `mapstructure:"preserve_host"`
	Headers       map[string]string `mapstructure:"headers"`
	Timeout       time.Duration     `mapstructure:"timeout"`
	RetryCount    int               `mapstructure:"retry_count"`
	LoadBalancing string            `mapstructure:"load_balancing"`
	HealthCheck   *HealthCheckConfig `mapstructure:"health_check"`
	Authentication *AuthConfig      `mapstructure:"authentication"`
	RateLimit     *RateLimitConfig  `mapstructure:"rate_limit"`
	Compression   bool              `mapstructure:"compression"`
}

type HealthCheckConfig struct {
	Path     string        `mapstructure:"path"`
	Interval time.Duration `mapstructure:"interval"`
	Timeout  time.Duration `mapstructure:"timeout"`
	Healthy  int           `mapstructure:"healthy"`
	Unhealthy int          `mapstructure:"unhealthy"`
}

type AuthConfig struct {
	Type    string            `mapstructure:"type"`
	Options map[string]string `mapstructure:"options"`
}

type RateLimitConfig struct {
	Rate     float64 `mapstructure:"rate"`
	Burst    int     `mapstructure:"burst"`
	Strategy string  `mapstructure:"strategy"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
	File   string `mapstructure:"file"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

type PluginsConfig struct {
	Dir     string   `mapstructure:"dir"`
	Enabled []string `mapstructure:"enabled"`
}

type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
	MinVersion string `mapstructure:"min_version"`
	AutoCert  bool   `mapstructure:"auto_cert"`
	AutoCertDir string `mapstructure:"auto_cert_dir"`
}

type HealthConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

type SecurityConfig struct {
	MaxBodySize    int64    `mapstructure:"max_body_size"`
	AllowedHosts   []string `mapstructure:"allowed_hosts"`
	TrustedProxies []string `mapstructure:"trusted_proxies"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     120 * time.Second,
			MaxHeaderBytes:  1 << 20,
			ShutdownTimeout: 10 * time.Second,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Port:    9090,
			Path:    "/metrics",
		},
		Health: HealthConfig{
			Enabled: true,
			Path:    "/health",
		},
		Plugins: PluginsConfig{
			Dir: "./plugins",
		},
		Security: SecurityConfig{
			MaxBodySize: 10 << 20,
		},
	}
}

func LoadConfig(configFile string) (*Config, error) {
	v := viper.New()

	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("gatewayx")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/gatewayx/")
	}

	v.SetEnvPrefix("GATEWAYX")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	cfg := DefaultConfig()

	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	for i := range cfg.Routes {
		if cfg.Routes[i].LoadBalancing == "" {
			cfg.Routes[i].LoadBalancing = "round_robin"
		}
		if cfg.Routes[i].Timeout == 0 {
			cfg.Routes[i].Timeout = 30 * time.Second
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.TLS.Enabled {
		if !c.TLS.AutoCert {
			if c.TLS.CertFile == "" {
				return fmt.Errorf("TLS enabled but no cert_file provided")
			}
			if _, err := os.Stat(c.TLS.CertFile); err != nil {
				return fmt.Errorf("cert_file not found: %w", err)
			}
			if c.TLS.KeyFile == "" {
				return fmt.Errorf("TLS enabled but no key_file provided")
			}
			if _, err := os.Stat(c.TLS.KeyFile); err != nil {
				return fmt.Errorf("key_file not found: %w", err)
			}
		}
	}

	for i, route := range c.Routes {
		if route.ListenPath == "" {
			return fmt.Errorf("route %d: listen_path is required", i)
		}
		if len(route.UpstreamURLs) == 0 {
			return fmt.Errorf("route %d (%s): at least one upstream_url is required", i, route.Name)
		}
	}

	return nil
}
