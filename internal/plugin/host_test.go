package plugin

import (
	"context"
	"net/http"
	"os"
	"testing"

	sdkplugin "github.com/oni1997/gatewayx/sdk/plugin"
)

type testPlugin struct {
	name       string
	version    string
	initErr    error
	reqErr     error
	initCalled bool
}

func (p *testPlugin) Name() string    { return p.name }
func (p *testPlugin) Version() string { return p.version }
func (p *testPlugin) Init(config map[string]any) error {
	p.initCalled = true
	return p.initErr
}
func (p *testPlugin) OnRequest(ctx context.Context, req *http.Request) (context.Context, error) {
	return ctx, p.reqErr
}
func (p *testPlugin) OnResponse(ctx context.Context, res *http.Response) error { return nil }
func (p *testPlugin) OnError(ctx context.Context, err error)                   {}
func (p *testPlugin) Close() error                                             { return nil }

func TestHost_LoadAndUnload(t *testing.T) {
	host := NewHost(sdkplugin.NewEvents())
	p := &testPlugin{name: "test", version: "1.0"}

	if err := host.Load(p, nil); err != nil {
		t.Fatalf("failed to load plugin: %v", err)
	}

	if !p.initCalled {
		t.Error("expected Init to be called")
	}

	loaded, ok := host.Get("test")
	if !ok {
		t.Fatal("plugin not found in host")
	}
	if loaded.Name() != "test" {
		t.Errorf("expected test, got %s", loaded.Name())
	}

	if err := host.Unload("test"); err != nil {
		t.Fatalf("failed to unload: %v", err)
	}

	if _, ok := host.Get("test"); ok {
		t.Error("plugin should be unloaded")
	}
}

func TestHost_DuplicateLoad(t *testing.T) {
	host := NewHost(sdkplugin.NewEvents())
	p := &testPlugin{name: "dup", version: "1.0"}

	_ = host.Load(p, nil)
	err := host.Load(p, nil)
	if err == nil {
		t.Error("expected error on duplicate load")
	}
}

func TestRegistry_RegisterAndList(t *testing.T) {
	reg := sdkplugin.NewRegistry(sdkplugin.NewEvents())

	p1 := &testPlugin{name: "p1", version: "1.0"}
	p2 := &testPlugin{name: "p2", version: "2.0"}

	_ = reg.Register(p1)
	_ = reg.Register(p2)

	if reg.Count() != 2 {
		t.Errorf("expected 2 plugins, got %d", reg.Count())
	}

	info := reg.List()
	if len(info) != 2 {
		t.Errorf("expected 2 in list, got %d", len(info))
	}
}

func TestRegistry_Unregister(t *testing.T) {
	reg := sdkplugin.NewRegistry(sdkplugin.NewEvents())
	_ = reg.Register(&testPlugin{name: "p1", version: "1.0"})
	_ = reg.Register(&testPlugin{name: "p2", version: "1.0"})

	reg.Unregister("p1")
	if reg.Count() != 1 {
		t.Errorf("expected 1 after unregister, got %d", reg.Count())
	}
}

func TestEvents_RegisterAndTrigger(t *testing.T) {
	events := sdkplugin.NewEvents()

	called := false
	events.Register(sdkplugin.HookOnStart, func(ctx context.Context, data map[string]any) (context.Context, error) {
		called = true
		return ctx, nil
	})

	if !events.HasHooks(sdkplugin.HookOnStart) {
		t.Error("expected hooks registered for OnStart")
	}

	ctx := context.Background()
	_, err := events.Trigger(sdkplugin.HookOnStart, ctx, nil)
	if err != nil {
		t.Fatalf("trigger failed: %v", err)
	}
	if !called {
		t.Error("hook should have been called")
	}
}

func TestEvents_ErrorPropagation(t *testing.T) {
	events := sdkplugin.NewEvents()

	events.Register(sdkplugin.HookPreRequest, func(ctx context.Context, data map[string]any) (context.Context, error) {
		return ctx, context.DeadlineExceeded
	})

	ctx := context.Background()
	_, err := events.Trigger(sdkplugin.HookPreRequest, ctx, nil)
	if err == nil {
		t.Error("expected error propagation")
	}
}

func TestPluginConfig_LoadConfig(t *testing.T) {
	configJSON := `[
		{"name": "auth-jwt", "enabled": true, "config": {"secret": "abc"}},
		{"name": "monitoring", "enabled": false, "config": {}}
	]`

	dir := t.TempDir()
	path := dir + "/plugins.json"
	if err := os.WriteFile(path, []byte(configJSON), 0644); err != nil {
		t.Fatal(err)
	}

	configs, err := sdkplugin.LoadConfig(path)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if len(configs) != 2 {
		t.Errorf("expected 2 configs, got %d", len(configs))
	}
	if configs[0].Name != "auth-jwt" {
		t.Errorf("expected auth-jwt, got %s", configs[0].Name)
	}
}

func TestScanDir(t *testing.T) {
	dir := t.TempDir()

	_ = os.WriteFile(dir+"/plugin.so", []byte("mock"), 0644)
	_ = os.WriteFile(dir+"/not-a-plugin.txt", []byte("text"), 0644)

	files, err := sdkplugin.ScanDir(dir)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 .so file, got %d", len(files))
	}
}

func TestManager_MiddlewarePasses(t *testing.T) {
	host := NewHost(sdkplugin.NewEvents())
	events := sdkplugin.NewEvents()
	mgr := NewManager(host, events)

	_ = host.Load(&testPlugin{name: "ok-plugin", version: "1.0"}, nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := mgr.Middleware(handler)

	req, _ := http.NewRequest("GET", "/test", nil)
	rec := newResponseRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.status != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.status)
	}
}

func TestManager_MiddlewareBlocksError(t *testing.T) {
	host := NewHost(sdkplugin.NewEvents())
	events := sdkplugin.NewEvents()
	mgr := NewManager(host, events)

	_ = host.Load(&testPlugin{name: "err-plugin", version: "1.0", reqErr: context.DeadlineExceeded}, nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	wrapped := mgr.Middleware(handler)

	req, _ := http.NewRequest("GET", "/test", nil)
	rec := newResponseRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.status != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.status)
	}
}

type responseRecorder struct {
	headers http.Header
	status  int
	body    []byte
}

func newResponseRecorder() *responseRecorder {
	return &responseRecorder{headers: make(http.Header)}
}

func (r *responseRecorder) Header() http.Header { return r.headers }
func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return len(b), nil
}
func (r *responseRecorder) WriteHeader(code int) { r.status = code }
