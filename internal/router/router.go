package router

import (
	"net/http"
	"strings"
)

type Route struct {
	Methods   []string
	Paths     []string
	Hosts     []string
	Handler   http.Handler
	StripPath string
}

type Router struct {
	routes []Route
}

func New() *Router {
	return &Router{}
}

func (r *Router) Add(route Route) {
	r.routes = append(r.routes, route)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range r.routes {
		if !matchMethods(route.Methods, req.Method) {
			continue
		}
		if !matchHosts(route.Hosts, req.Host) {
			continue
		}
		if matched, strip := matchPath(route.Paths, req.URL.Path, route.StripPath); matched {
			if strip != "" && strings.HasPrefix(req.URL.Path, strip) {
				req.URL.Path = strings.TrimPrefix(req.URL.Path, strip)
				if req.URL.Path == "" {
					req.URL.Path = "/"
				}
			}
			route.Handler.ServeHTTP(w, req)
			return
		}
	}
	http.NotFound(w, req)
}

func matchMethods(allowed []string, method string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, m := range allowed {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

func matchHosts(allowed []string, host string) bool {
	if len(allowed) == 0 {
		return true
	}
	h := strings.Split(host, ":")[0]
	for _, a := range allowed {
		if strings.EqualFold(a, h) {
			return true
		}
	}
	return false
}

func matchPath(paths []string, reqPath string, stripPath string) (bool, string) {
	if len(paths) == 0 {
		return true, stripPath
	}
	for _, p := range paths {
		if strings.HasPrefix(reqPath, p) {
			return true, p
		}
	}
	return false, ""
}
