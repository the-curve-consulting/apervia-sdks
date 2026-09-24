package apervia

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromContextIsFalseWithoutPlatformHeaders(t *testing.T) {
	var got bool
	handler := Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, got = FromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if got {
		t.Fatal("FromContext second value = true, want false")
	}
}

func TestMiddlewareReadsAppIdentity(t *testing.T) {
	var id Identity
	var ok bool
	handler := Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		id, ok = FromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Apervia-User-ID", "user-1")
	req.Header.Set("X-Apervia-Tenant-ID", "acme")
	req.Header.Set("X-Apervia-Token-ID", "tok-1")
	req.Header.Set("X-Apervia-Env", "production")
	req.Header.Set("X-Apervia-App-Permissions", " finance:read , finance:write, ")
	req.Header.Set("X-Apervia-App-Roles", "finance-manager, auditor")
	req.Header.Set("X-Apervia-Roles", "platform-admin")
	req.Header.Set("X-Apervia-Permissions", "apps:manage")
	req.Header.Set("X-Platform-User-ID", "forged")

	handler.ServeHTTP(httptest.NewRecorder(), req)
	if !ok {
		t.Fatal("FromContext second value = false, want true")
	}
	if id.UserID != "user-1" || id.TenantID != "acme" || id.TokenID != "tok-1" || id.Environment != "production" {
		t.Fatalf("identity = %+v", id)
	}
	if !id.Can("finance:read") || !id.Can("finance:write") || id.Can("apps:manage") {
		t.Fatalf("permissions = %v", id.Permissions)
	}
	if !id.HasRole("finance-manager") || !id.HasRole("auditor") || id.HasRole("platform-admin") {
		t.Fatalf("roles = %v", id.Roles)
	}
}

func TestEmptyPermissionsAreNotAnonymous(t *testing.T) {
	var id Identity
	var ok bool
	handler := Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		id, ok = FromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Apervia-User-ID", "user-1")
	req.Header.Set("X-Apervia-Tenant-ID", "acme")
	req.Header.Set("X-Apervia-Token-ID", "tok-1")
	req.Header.Set("X-Apervia-Env", "production")
	req.Header.Set("X-Apervia-App-Permissions", "")
	req.Header.Set("X-Apervia-App-Roles", "finance-manager")

	handler.ServeHTTP(httptest.NewRecorder(), req)
	if !ok {
		t.Fatal("empty permissions were treated as anonymous")
	}
	if len(id.Permissions) != 0 || !id.HasRole("finance-manager") {
		t.Fatalf("identity = %+v", id)
	}
}

func TestOldPrefixOnlyIsAnonymous(t *testing.T) {
	var ok bool
	handler := Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, ok = FromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Platform-User-ID", "user-1")
	req.Header.Set("X-Platform-Tenant-ID", "acme")
	req.Header.Set("X-Platform-Token-ID", "tok-1")

	handler.ServeHTTP(httptest.NewRecorder(), req)
	if ok {
		t.Fatal("old prefix was treated as a platform identity")
	}
}

func TestWithDevelopmentIdentityOverridesHeaders(t *testing.T) {
	dev := Identity{UserID: "dev", TenantID: "local", TokenID: "dev-token", Permissions: []string{"demo:admin"}}
	var id Identity
	var ok bool
	handler := Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		id, ok = FromContext(r.Context())
	}), WithDevelopmentIdentity(dev))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Apervia-User-ID", "real-user")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if !ok || id.UserID != "dev" || !id.Can("demo:admin") {
		t.Fatalf("identity = %+v ok=%v", id, ok)
	}
}
