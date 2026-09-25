// Package aperviatest signs a test request in as a named persona.
// Import it from test files only. It sets real platform headers, and the
// request then passes through apervia.Middleware. It never uses
// WithDevelopmentIdentity.
//
// Declare Personas once, in TestMain. A second call panics.
//
//	func TestMain(m *testing.M) {
//		aperviatest.Personas(map[string]aperviatest.Persona{
//			"alice": {
//				Permissions: []string{"finance:read"},
//				Roles:       []string{"finance-manager"},
//				TenantID:    "acme",
//			},
//		})
//		os.Exit(m.Run())
//	}
//
//	func TestWhoami(t *testing.T) {
//		req := aperviatest.LoginAs(httptest.NewRequest(http.MethodGet, "/", nil), "alice")
//		var id apervia.Identity
//		apervia.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
//			id, _ = apervia.FromContext(r.Context())
//		})).ServeHTTP(httptest.NewRecorder(), req)
//	}
package aperviatest

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/the-curve-consulting/apervia-sdks/go"
)

// personaNamespace is the UUID v5 namespace shared by every SDK test helper.
// The persona name is the UUID name. See conformance/personas.json.
const personaNamespace = "6f0b5c2e-1a4d-4e7b-9c3a-8d2e5f7a1b04"

const (
	headerUserID      = "X-Apervia-User-ID"
	headerTenantID    = "X-Apervia-Tenant-ID"
	headerTokenID     = "X-Apervia-Token-ID"
	headerEnvironment = "X-Apervia-Env"
	headerPermissions = "X-Apervia-App-Permissions"
	headerRoles       = "X-Apervia-App-Roles"
)

// Persona is one app user. Environment may be omitted. TenantID may be
// omitted only for a name used with UserID. LoginAs requires it.
type Persona struct {
	Permissions []string
	Roles       []string
	TenantID    string
	Environment string
}

var (
	registryMu sync.Mutex
	registry   map[string]Persona
	frozen     bool
)

// Personas declares the app's personas once. A second call panics.
func Personas(personas map[string]Persona) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if frozen {
		panic("aperviatest.Personas called more than once")
	}
	registry = make(map[string]Persona, len(personas))
	for name, persona := range personas {
		persona.Permissions = copyList(persona.Permissions)
		persona.Roles = copyList(persona.Roles)
		registry[name] = persona
	}
	frozen = true
}

// UserID returns the deterministic user id for a persona name.
func UserID(name string) string {
	return uuidV5(personaNamespace, name)
}

// Identity returns the platform identity for a declared persona.
func Identity(name string) apervia.Identity {
	persona := lookup(name)
	return apervia.Identity{
		UserID:      UserID(name),
		TenantID:    persona.TenantID,
		TokenID:     "test-" + name,
		Environment: persona.Environment,
		Permissions: copyList(persona.Permissions),
		Roles:       copyList(persona.Roles),
	}
}

// LoginAs sets the platform headers for a declared persona.
func LoginAs(req *http.Request, name string) *http.Request {
	return LoginAsIdentity(req, Identity(name))
}

// LoginAsIdentity sets the platform headers for a one-off identity.
// It panics when UserID, TenantID, or TokenID is blank, because Middleware
// would treat that request as anonymous.
func LoginAsIdentity(req *http.Request, id apervia.Identity) *http.Request {
	switch {
	case strings.TrimSpace(id.UserID) == "":
		panic("aperviatest: UserID is empty; Middleware would treat the request as anonymous")
	case strings.TrimSpace(id.TenantID) == "":
		panic("aperviatest: TenantID is empty; Middleware would treat the request as anonymous")
	case strings.TrimSpace(id.TokenID) == "":
		panic("aperviatest: TokenID is empty; Middleware would treat the request as anonymous")
	}
	req.Header.Set(headerUserID, id.UserID)
	req.Header.Set(headerTenantID, id.TenantID)
	req.Header.Set(headerTokenID, id.TokenID)
	req.Header.Set(headerEnvironment, id.Environment)
	req.Header.Set(headerPermissions, strings.Join(id.Permissions, ","))
	req.Header.Set(headerRoles, strings.Join(id.Roles, ","))
	return req
}

// Anonymous removes the platform identity headers.
func Anonymous(req *http.Request) *http.Request {
	for _, name := range []string{
		headerUserID, headerTenantID, headerTokenID,
		headerEnvironment, headerPermissions, headerRoles,
	} {
		req.Header.Del(name)
	}
	return req
}

func lookup(name string) Persona {
	registryMu.Lock()
	defer registryMu.Unlock()
	persona, ok := registry[name]
	if !ok {
		panic(fmt.Sprintf("aperviatest: unknown persona %q", name))
	}
	return persona
}

func copyList(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	out := make([]string, len(items))
	copy(out, items)
	return out
}

func uuidV5(namespace, name string) string {
	ns := mustParseUUID(namespace)
	sum := sha1.Sum(append(ns[:], name...))
	var id [16]byte
	copy(id[:], sum[:16])
	id[6] = (id[6] & 0x0f) | 0x50
	id[8] = (id[8] & 0x3f) | 0x80
	return formatUUID(id)
}

func mustParseUUID(value string) [16]byte {
	raw, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	if err != nil || len(raw) != 16 {
		panic("aperviatest: invalid namespace uuid")
	}
	var id [16]byte
	copy(id[:], raw)
	return id
}

func formatUUID(id [16]byte) string {
	encoded := hex.EncodeToString(id[:])
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32]
}
