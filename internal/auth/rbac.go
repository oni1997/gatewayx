package auth

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

type Permission struct {
	Path    string
	Methods []string
	Roles   []string
}

type Role struct {
	Name        string
	Permissions []string
}

type RBACEngine struct {
	permissions []Permission
	roles       map[string]Role
}

func NewRBACEngine() *RBACEngine {
	return &RBACEngine{
		roles: make(map[string]Role),
	}
}

func (rbac *RBACEngine) AddRole(role Role) {
	rbac.roles[role.Name] = role
}

func (rbac *RBACEngine) AddPermission(perm Permission) {
	rbac.permissions = append(rbac.permissions, perm)
}

func (rbac *RBACEngine) CheckPermission(userRoles []string, method, path string) bool {
	for _, perm := range rbac.permissions {
		if !matchGlob(perm.Path, path) {
			continue
		}

		if len(perm.Methods) > 0 && !containsStringFold(perm.Methods, method) {
			continue
		}

		for _, userRole := range userRoles {
			if containsStringFold(perm.Roles, userRole) {
				return true
			}
		}
	}

	return false
}

func matchGlob(pattern, path string) bool {
	if idx := strings.Index(pattern, "**"); idx >= 0 {
		prefix := pattern[:idx]
		prefix = strings.TrimSuffix(prefix, "/")
		if prefix == "" || prefix == "/" {
			return true
		}
		pathParts := cleanPathParts(path)
		prefixParts := cleanPathParts(prefix)
		if len(pathParts) < len(prefixParts) {
			return false
		}
		pathPrefix := strings.Join(pathParts[:len(prefixParts)], "/")
		matched, _ := filepath.Match(strings.Join(prefixParts, "/"), pathPrefix)
		return matched
	}

	matched, _ := filepath.Match(pattern, path)
	if matched {
		return true
	}

	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		rest := strings.TrimPrefix(path, prefix+"/")
		return strings.HasPrefix(path, prefix+"/") && !strings.Contains(rest, "/")
	}

	return false
}

func cleanPathParts(p string) []string {
	parts := strings.Split(p, "/")
	var cleaned []string
	for _, part := range parts {
		if part != "" {
			cleaned = append(cleaned, part)
		}
	}
	return cleaned
}

func containsStringFold(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

type RBACAuthenticator struct {
	delegate   Authenticator
	engine     *RBACEngine
	rolesClaim string
}

func NewRBAC(delegate Authenticator, engine *RBACEngine, rolesClaim string) *RBACAuthenticator {
	if rolesClaim == "" {
		rolesClaim = "roles"
	}
	return &RBACAuthenticator{
		delegate:   delegate,
		engine:     engine,
		rolesClaim: rolesClaim,
	}
}

func (ra *RBACAuthenticator) Name() string {
	return "rbac"
}

func (ra *RBACAuthenticator) Authenticate(r *http.Request) (Claims, error) {
	claims, err := ra.delegate.Authenticate(r)
	if err != nil {
		return nil, err
	}

	rolesRaw, ok := claims[ra.rolesClaim]
	if !ok {
		return nil, fmt.Errorf("no roles claim found in token")
	}

	var userRoles []string
	switch v := rolesRaw.(type) {
	case []any:
		for _, r := range v {
			if s, ok := r.(string); ok {
				userRoles = append(userRoles, s)
			}
		}
	case []string:
		userRoles = v
	case string:
		userRoles = strings.Split(v, ",")
		for i := range userRoles {
			userRoles[i] = strings.TrimSpace(userRoles[i])
		}
	default:
		return nil, fmt.Errorf("unexpected roles type: %T", rolesRaw)
	}

	if len(userRoles) == 0 {
		return nil, fmt.Errorf("user has no assigned roles")
	}

	claims["_roles"] = userRoles
	claims["_rbac"] = true

	return claims, nil
}

func (ra *RBACAuthenticator) CheckAccess(userRoles []string, method, path string) bool {
	return ra.engine.CheckPermission(userRoles, method, path)
}

func RBACMiddleware(ra *RBACAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaims(r.Context())
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"unauthorized","message":"no auth claims in context"}`))
				return
			}

			rolesRaw, ok := claims["_roles"]
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden","message":"no roles in request context"}`))
				return
			}

			var userRoles []string
			switch v := rolesRaw.(type) {
			case []string:
				userRoles = v
			case []any:
				for _, r := range v {
					if s, ok := r.(string); ok {
						userRoles = append(userRoles, s)
					}
				}
			default:
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			if !ra.engine.CheckPermission(userRoles, r.Method, r.URL.Path) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden","message":"insufficient permissions"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
