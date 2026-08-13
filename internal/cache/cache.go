package cache

import (
	"bytes"
	"net/http"
	"sync"
	"time"
)

type entry struct {
	status    int
	header    http.Header
	body      []byte
	expiresAt time.Time
}

type Cache struct {
	mu      sync.RWMutex
	entries map[string]*entry
	ttl     time.Duration
	maxSize int
}

func New(ttl time.Duration, maxSize int) *Cache {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &Cache{
		entries: make(map[string]*entry),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

func (c *Cache) key(r *http.Request) string {
	return r.Method + ":" + r.URL.Path + "?" + r.URL.RawQuery
}

func (c *Cache) Get(r *http.Request) (*entry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.entries[c.key(r)]
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expiresAt) {
		delete(c.entries, c.key(r))
		return nil, false
	}
	return e, true
}

func (c *Cache) Set(r *http.Request, status int, header http.Header, body []byte) {
	if len(body) > 1024*1024 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= c.maxSize {
		for k := range c.entries {
			delete(c.entries, k)
			break
		}
	}

	c.entries[c.key(r)] = &entry{
		status:    status,
		header:    header.Clone(),
		body:      append([]byte(nil), body...),
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

func (c *Cache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*entry)
}

func Middleware(c *Cache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				next.ServeHTTP(w, r)
				return
			}

			if e, ok := c.Get(r); ok {
				for k, v := range e.header {
					for _, vv := range v {
						w.Header().Add(k, vv)
					}
				}
				w.Header().Set("X-Cache", "HIT")
				w.WriteHeader(e.status)
				_, _ = w.Write(e.body)
				return
			}

			rec := &recorder{header: make(http.Header)}
			next.ServeHTTP(rec, r)

			if rec.status >= 200 && rec.status < 300 {
				c.Set(r, rec.status, rec.header, rec.body.Bytes())
			}

			for k, v := range rec.header {
				for _, vv := range v {
					w.Header().Add(k, vv)
				}
			}
			w.Header().Set("X-Cache", "MISS")
			w.WriteHeader(rec.status)
			_, _ = w.Write(rec.body.Bytes())
		})
	}
}

type recorder struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (r *recorder) Header() http.Header         { return r.header }
func (r *recorder) Write(b []byte) (int, error) { return r.body.Write(b) }
func (r *recorder) WriteHeader(code int)        { r.status = code }
