package apervia_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/the-curve-consulting/apervia-sdks/go"
)

func ExampleMiddleware() {
	handler := apervia.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		id, ok := apervia.FromContext(r.Context())
		if !ok {
			return
		}
		fmt.Println(id.UserID, id.Can("finance:read"), id.HasRole("finance-manager"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Apervia-User-ID", "user-1")
	req.Header.Set("X-Apervia-Tenant-ID", "acme")
	req.Header.Set("X-Apervia-Token-ID", "tok-1")
	req.Header.Set("X-Apervia-App-Permissions", "finance:read")
	req.Header.Set("X-Apervia-App-Roles", "finance-manager")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	// Output:
	// user-1 true true
}
