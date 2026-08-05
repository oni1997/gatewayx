package plugin

import (
	"context"
	"net/http"
)

type Plugin interface {
	Name() string
	Version() string
	Init(config map[string]any) error
	OnRequest(ctx context.Context, req *http.Request) (context.Context, error)
	OnResponse(ctx context.Context, res *http.Response) error
	OnError(ctx context.Context, err error)
	Close() error
}

type LifecycleState string

const (
	StateUnloaded    LifecycleState = "unloaded"
	StateLoaded      LifecycleState = "loaded"
	StateInitialized LifecycleState = "initialized"
	StateRunning     LifecycleState = "running"
	StateStopped     LifecycleState = "stopped"
	StateError       LifecycleState = "error"
)

type PluginInfo struct {
	Name    string            `json:"name"`
	Version string            `json:"version"`
	State   LifecycleState    `json:"state"`
	Config  map[string]any    `json:"config,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

type Hook string

const (
	HookPreRequest   Hook = "pre_request"
	HookPostRequest  Hook = "post_request"
	HookPreResponse  Hook = "pre_response"
	HookPostResponse Hook = "post_response"
	HookOnError      Hook = "on_error"
	HookOnStart      Hook = "on_start"
	HookOnStop       Hook = "on_stop"
)

type HookFunc func(ctx context.Context, data map[string]any) (context.Context, error)

type Events struct {
	hooks map[Hook][]HookFunc
}

func NewEvents() *Events {
	return &Events{hooks: make(map[Hook][]HookFunc)}
}

func (e *Events) Register(hook Hook, fn HookFunc) {
	e.hooks[hook] = append(e.hooks[hook], fn)
}

func (e *Events) Trigger(hook Hook, ctx context.Context, data map[string]any) (context.Context, error) {
	for _, fn := range e.hooks[hook] {
		var err error
		ctx, err = fn(ctx, data)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (e *Events) HasHooks(hook Hook) bool {
	return len(e.hooks[hook]) > 0
}
