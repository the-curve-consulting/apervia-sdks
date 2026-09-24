package apervia_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/mux"
	"github.com/the-curve-consulting/apervia-sdks/go"
)

func TestSameHandlerOnStdlibChiAndGorilla(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := apervia.FromContext(r.Context())
		if !ok || id.UserID != "user-1" || !id.Can("finance:read") {
			http.Error(w, "missing identity", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	stdlib := http.NewServeMux()
	stdlib.Handle("GET /whoami", handler)

	chiRouter := chi.NewRouter()
	chiRouter.Get("/whoami", handler.ServeHTTP)

	gorilla := mux.NewRouter()
	gorilla.Handle("/whoami", handler).Methods(http.MethodGet)

	for name, router := range map[string]http.Handler{
		"stdlib":  apervia.Middleware(stdlib),
		"chi":     apervia.Middleware(chiRouter),
		"gorilla": apervia.Middleware(gorilla),
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
			req.Header.Set("X-Apervia-User-ID", "user-1")
			req.Header.Set("X-Apervia-Tenant-ID", "acme")
			req.Header.Set("X-Apervia-Token-ID", "tok-1")
			req.Header.Set("X-Apervia-App-Permissions", "finance:read")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusNoContent {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}
		})
	}
}
