package apervia

import (
	"context"
	"net/http"
	"strings"
)

const (
	headerUserID      = "X-Apervia-User-ID"
	headerTenantID    = "X-Apervia-Tenant-ID"
	headerTokenID     = "X-Apervia-Token-ID"
	headerEnvironment = "X-Apervia-Env"
	headerPermissions = "X-Apervia-App-Permissions"
	headerRoles       = "X-Apervia-App-Roles"
)

// Identity is the app-plane caller the platform already authenticated.
// It has no system-plane roles or permissions.
type Identity struct {
	UserID      string
	TenantID    string
	TokenID     string
	Environment string
	Permissions []string
	Roles       []string
}

// Can reports whether the caller holds one app-plane permission.
func (id Identity) Can(permission string) bool {
	return contains(id.Permissions, permission)
}

// HasRole reports whether the caller holds one app-plane role name.
func (id Identity) HasRole(role string) bool {
	return contains(id.Roles, role)
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

type contextKey struct{}

// FromContext returns the identity stored by Middleware.
// The second value is false for an anonymous request.
func FromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(contextKey{}).(Identity)
	return id, ok
}

// Option configures Middleware.
type Option func(*middlewareConfig)

type middlewareConfig struct {
	development *Identity
}

// WithDevelopmentIdentity makes Middleware store id and ignore request headers.
// The middleware uses it only when the caller passes this option.
func WithDevelopmentIdentity(id Identity) Option {
	return func(cfg *middlewareConfig) {
		copied := id
		cfg.development = &copied
	}
}

// Middleware reads the platform identity headers and stores an Identity in the
// request context. A request is anonymous unless user, tenant, and token id
// are all present, so a partial set of headers is never returned as an identity.
func Middleware(next http.Handler, opts ...Option) http.Handler {
	var cfg middlewareConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if cfg.development != nil {
			ctx = context.WithValue(ctx, contextKey{}, *cfg.development)
		} else if id, ok := identityFrom(r.Header); ok {
			ctx = context.WithValue(ctx, contextKey{}, id)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func identityFrom(h http.Header) (Identity, bool) {
	userID := strings.TrimSpace(h.Get(headerUserID))
	tenantID := strings.TrimSpace(h.Get(headerTenantID))
	tokenID := strings.TrimSpace(h.Get(headerTokenID))
	if userID == "" || tenantID == "" || tokenID == "" {
		return Identity{}, false
	}
	return Identity{
		UserID:      userID,
		TenantID:    tenantID,
		TokenID:     tokenID,
		Environment: strings.TrimSpace(h.Get(headerEnvironment)),
		Permissions: splitList(h.Get(headerPermissions)),
		Roles:       splitList(h.Get(headerRoles)),
	}, true
}

func splitList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}
