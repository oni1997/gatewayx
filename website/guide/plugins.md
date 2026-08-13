# Plugins

GatewayX supports a plugin system to extend its functionality.

## Plugin interface

```go
type Plugin interface {
    Name() string
    Version() string
    Init(config map[string]any) error
    OnRequest(ctx context.Context, req *http.Request) (context.Context, error)
    OnResponse(ctx context.Context, res *http.Response) error
    OnError(ctx context.Context, err error)
    Close() error
}
```

## Building a plugin

```bash
go build -buildmode=plugin -o plugins/myplugin.so ./path/to/plugin
```

See [Plugins](/guide/plugins) and the `sdk/` directory for details.
