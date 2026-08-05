package plugin

import (
	"fmt"
	"sync"

	sdkplugin "github.com/oni1997/gatewayx/sdk/plugin"
)

type Host struct {
	mu      sync.RWMutex
	plugins map[string]sdkplugin.Plugin
	events  *sdkplugin.Events
}

func NewHost(events *sdkplugin.Events) *Host {
	return &Host{
		plugins: make(map[string]sdkplugin.Plugin),
		events:  events,
	}
}

func (h *Host) Load(p sdkplugin.Plugin, config map[string]any) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.plugins[p.Name()]; exists {
		return fmt.Errorf("plugin %s is already loaded", p.Name())
	}

	if err := p.Init(config); err != nil {
		return fmt.Errorf("plugin %s init failed: %w", p.Name(), err)
	}

	h.plugins[p.Name()] = p
	return nil
}

func (h *Host) Unload(name string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	p, exists := h.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not loaded", name)
	}

	if err := p.Close(); err != nil {
		return fmt.Errorf("plugin %s close failed: %w", name, err)
	}

	delete(h.plugins, name)
	return nil
}

func (h *Host) Get(name string) (sdkplugin.Plugin, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	p, ok := h.plugins[name]
	return p, ok
}

func (h *Host) List() []sdkplugin.Plugin {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]sdkplugin.Plugin, 0, len(h.plugins))
	for _, p := range h.plugins {
		result = append(result, p)
	}
	return result
}

func (h *Host) Shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for name, p := range h.plugins {
		p.Close()
		delete(h.plugins, name)
	}
}
