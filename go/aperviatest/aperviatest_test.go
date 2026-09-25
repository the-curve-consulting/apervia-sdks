package aperviatest

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/the-curve-consulting/apervia-sdks/go"
)

func TestMain(m *testing.M) {
	if os.Getenv("APERVIATEST_EXPECT_PANIC") == "1" {
		os.Exit(m.Run())
	}
	Personas(map[string]Persona{
		"alice": {
			Permissions: []string{"finance:read", "finance:write"},
			Roles:       []string{"finance-manager"},
			TenantID:    "acme",
			Environment: "production",
		},
		"guest": {},
	})
	os.Exit(m.Run())
}

func TestUserIDMatchesTheFixture(t *testing.T) {
	if got := UserID("alice"); got != "75431223-00e0-5f1b-ac66-bbfba81540cd" {
		t.Fatalf("UserID(alice) = %s", got)
	}
	if got := UserID("bob"); got != "2e837302-bf97-50fe-a215-cf2f3320432a" {
		t.Fatalf("UserID(bob) = %s", got)
	}
	if got := UserID("guest"); got != "8bea96e2-4c0d-5027-be1c-e9edfbee95e3" {
		t.Fatalf("UserID(guest) = %s", got)
	}
}

func TestLoginAsGoesThroughMiddleware(t *testing.T) {
	var id apervia.Identity
	var ok bool
	handler := apervia.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		id, ok = apervia.FromContext(r.Context())
	}))

	req := LoginAs(httptest.NewRequest(http.MethodGet, "/", nil), "alice")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if !ok || id.UserID != UserID("alice") || id.TenantID != "acme" || id.Environment != "production" {
		t.Fatalf("identity = %+v ok=%v", id, ok)
	}
	if !id.Can("finance:read") || !id.HasRole("finance-manager") {
		t.Fatalf("identity = %+v", id)
	}
}

func TestAnonymousRemovesPlatformHeaders(t *testing.T) {
	req := LoginAs(httptest.NewRequest(http.MethodGet, "/", nil), "alice")
	req = Anonymous(req)

	var ok bool
	apervia.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, ok = apervia.FromContext(r.Context())
	})).ServeHTTP(httptest.NewRecorder(), req)
	if ok {
		t.Fatal("anonymous request was signed in")
	}
}

func TestLoginAsIdentity(t *testing.T) {
	want := apervia.Identity{
		UserID: "one-off", TenantID: "acme", TokenID: "tok",
		Permissions: []string{"demo:admin"}, Roles: []string{"auditor"},
	}
	var got apervia.Identity
	var ok bool
	apervia.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got, ok = apervia.FromContext(r.Context())
	})).ServeHTTP(httptest.NewRecorder(), LoginAsIdentity(httptest.NewRequest(http.MethodGet, "/", nil), want))
	if !ok || got.UserID != "one-off" || !got.Can("demo:admin") || !got.HasRole("auditor") {
		t.Fatalf("identity = %+v ok=%v", got, ok)
	}
}

func TestLoginAsIdentityPanicsWithoutTenant(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("blank TenantID did not panic")
		}
	}()
	LoginAsIdentity(httptest.NewRequest(http.MethodGet, "/", nil), apervia.Identity{
		UserID: "user-1", TokenID: "tok-1",
	})
}

func TestSecondPersonasCallPanics(t *testing.T) {
	if os.Getenv("APERVIATEST_EXPECT_PANIC") == "1" {
		Personas(map[string]Persona{"alice": {}})
		Personas(map[string]Persona{"bob": {}})
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestSecondPersonasCallPanics")
	cmd.Env = append(os.Environ(), "APERVIATEST_EXPECT_PANIC=1")
	if err := cmd.Run(); err == nil {
		t.Fatal("second Personas call did not panic")
	}
}
