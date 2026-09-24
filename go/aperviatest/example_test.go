package aperviatest

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/the-curve-consulting/apervia-sdks/go"
)

func ExampleLoginAs() {
	var id apervia.Identity
	handler := apervia.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		id, _ = apervia.FromContext(r.Context())
	}))
	req := LoginAs(httptest.NewRequest(http.MethodGet, "/", nil), "alice")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	fmt.Println(id.UserID)
	// Output:
	// 75431223-00e0-5f1b-ac66-bbfba81540cd
}
