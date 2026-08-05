package example

import (
	"context"
	"fmt"
	"net/http"
	"os"

	plugin "github.com/oni1997/gatewayx/sdk/plugin"
)

var (
	PluginName    = "example-auth"
	PluginVersion = "1.0.0"
	apiKey        = os.Getenv("EXAMPLE_AUTH_KEY")
)

type ExampleAuthPlugin struct {
	config map[string]any
}

func (p *ExampleAuthPlugin) Name() string {
	return PluginName
}

func (p *ExampleAuthPlugin) Version() string {
	return PluginVersion
}

func (p *ExampleAuthPlugin) Init(config map[string]any) error {
	p.config = config
	if key, ok := config["api_key"].(string); ok && key != "" {
		apiKey = key
	}
	if apiKey == "" {
		return fmt.Errorf("example-auth: api_key not configured")
	}
	return nil
}

func (p *ExampleAuthPlugin) OnRequest(ctx context.Context, req *http.Request) (context.Context, error) {
	key := req.Header.Get("X-Plugin-Auth")
	if key == "" {
		return ctx, fmt.Errorf("missing X-Plugin-Auth header")
	}
	if key != apiKey {
		return ctx, fmt.Errorf("invalid plugin auth key")
	}
	return ctx, nil
}

func (p *ExampleAuthPlugin) OnResponse(ctx context.Context, res *http.Response) error {
	res.Header.Set("X-Plugin", PluginName)
	return nil
}

func (p *ExampleAuthPlugin) OnError(ctx context.Context, err error) {
	fmt.Fprintf(os.Stderr, "[%s] error: %v\n", PluginName, err)
}

func (p *ExampleAuthPlugin) Close() error {
	return nil
}

var Plugin plugin.Plugin = &ExampleAuthPlugin{}
