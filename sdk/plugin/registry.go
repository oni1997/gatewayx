package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Registry struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
	events  *Events
}

func NewRegistry(events *Events) *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
		events:  events,
	}
}

func (r *Registry) Register(p Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[p.Name()]; exists {
		return fmt.Errorf("plugin %s already registered", p.Name())
	}
	r.plugins[p.Name()] = p
	return nil
}

func (r *Registry) Get(name string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.plugins[name]
	return p, ok
}

func (r *Registry) List() []PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]PluginInfo, 0, len(r.plugins))
	for _, p := range r.plugins {
		result = append(result, PluginInfo{
			Name:    p.Name(),
			Version: p.Version(),
			State:   StateRunning,
		})
	}
	return result
}

func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p, ok := r.plugins[name]; ok {
		p.Close()
		delete(r.plugins, name)
	}
}

func (r *Registry) Shutdown() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, p := range r.plugins {
		p.Close()
		delete(r.plugins, name)
	}
}

func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.plugins)
}

type PluginConfig struct {
	Name    string         `json:"name"`
	Enabled bool           `json:"enabled"`
	Config  map[string]any `json:"config"`
}

func LoadConfig(path string) ([]PluginConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin config: %w", err)
	}

	var configs []PluginConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse plugin config: %w", err)
	}
	return configs, nil
}

func ScanDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var plugins []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".so") {
			plugins = append(plugins, filepath.Join(dir, name))
		}
	}
	return plugins, nil
}
