package apervia

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

type conformanceFile struct {
	Cases []conformanceCase `json:"cases"`
}

type conformanceCase struct {
	Name     string            `json:"name"`
	Headers  map[string]string `json:"headers"`
	Identity conformanceID     `json:"identity"`
}

type conformanceID struct {
	Anonymous   bool     `json:"anonymous"`
	UserID      string   `json:"userId"`
	TenantID    string   `json:"tenantId"`
	TokenID     string   `json:"tokenId"`
	Environment string   `json:"environment"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
}

func TestConformanceFixture(t *testing.T) {
	raw, err := os.ReadFile("../conformance/cases.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	var file conformanceFile
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("parsing fixture: %v", err)
	}
	if len(file.Cases) == 0 {
		t.Fatal("fixture has no cases")
	}

	for _, tc := range file.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			var id Identity
			var ok bool
			handler := Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				id, ok = FromContext(r.Context())
			}))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			for name, value := range tc.Headers {
				req.Header.Set(name, value)
			}
			handler.ServeHTTP(httptest.NewRecorder(), req)

			if tc.Identity.Anonymous {
				if ok {
					t.Fatalf("FromContext second value = true, want false, identity = %+v", id)
				}
				return
			}
			if !ok {
				t.Fatal("FromContext second value = false, want true")
			}
			got := conformanceID{
				UserID: id.UserID, TenantID: id.TenantID, TokenID: id.TokenID,
				Environment: id.Environment, Permissions: id.Permissions, Roles: id.Roles,
			}
			want := tc.Identity
			want.Anonymous = false
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("identity = %+v, want %+v", got, want)
			}
		})
	}
}
