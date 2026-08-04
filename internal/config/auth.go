package config

import (
	"os"
	"path/filepath"
)

func (ac *AuthConfig) JWTOptions() map[string]string {
	opts := make(map[string]string)
	for k, v := range ac.Options {
		val := os.ExpandEnv(v)
		opts[k] = val
	}
	return opts
}

func (ac *AuthConfig) APIKeyOptions() map[string]string {
	opts := make(map[string]string)
	for k, v := range ac.Options {
		val := os.ExpandEnv(v)
		if k == "keys_file" {
			if !filepath.IsAbs(val) {
				val = filepath.Join(".", val)
			}
		}
		opts[k] = val
	}
	return opts
}
