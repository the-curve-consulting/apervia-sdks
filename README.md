# apervia-sdks

SDKs that read the identity Apervia already authenticated. Each language lives in its own directory and releases on its own tag. Go, Ruby, and Node share one header contract and one conformance fixture.

The licence is Apache-2.0. See [LICENSE](LICENSE).

## Versioning

Each language releases independently:

| Language | Tag |
| --- | --- |
| Go | `go/v<x.y.z>` |
| Ruby | `ruby/v<x.y.z>` |
| Node | `node/v<x.y.z>` |

## Releasing

Run the Release workflow from `main`. Choose the language. For the first tag of that language, set version to `1.0.0`. After that, leave version empty and choose a patch, minor, or major bump.

The workflow pushes one tag, such as `go/v1.0.0`. That tag is the Go release. Nothing is uploaded. A Go major bump stays refused until the module path ends in `/v2`.

## Header contract

The sidecar sets these headers on every request that reaches an app. An SDK reads only these names:

| Header | Meaning |
| --- | --- |
| `X-Apervia-User-ID` | The user |
| `X-Apervia-Tenant-ID` | The tenant |
| `X-Apervia-Token-ID` | The token id, for audit correlation |
| `X-Apervia-Env` | The environment the app runs as |
| `X-Apervia-App-Permissions` | Comma-separated app-plane permissions |
| `X-Apervia-App-Roles` | Comma-separated app-plane role names |

The platform does not send `X-Apervia-App-Roles` yet. An SDK still accepts the header. An absent header is an empty role list.

The SDK ignores `X-Apervia-Roles` and `X-Apervia-Permissions`. Those names carry system-plane authority. The SDK also ignores every `X-Platform-*` header.

When the `X-Apervia-*` identity headers are absent, the SDK reports an anonymous request. It never returns a partial identity. An empty `X-Apervia-App-Permissions` header is an empty permission list, not an anonymous identity. An absent `X-Apervia-App-Roles` header is an empty role list.

The shared cases are in [conformance/cases.json](conformance/cases.json). Persona user ids are in [conformance/personas.json](conformance/personas.json).

## Go

```
go get github.com/the-curve-consulting/apervia-sdks/go
```

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
	id, ok := apervia.FromContext(r.Context())
	if !ok {
		http.Error(w, "anonymous", http.StatusUnauthorized)
		return
	}
	if !id.Can("finance:read") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	fmt.Fprintln(w, id.UserID)
})
http.ListenAndServe(":8080", apervia.Middleware(mux))
```

`apervia.Middleware` is an `http.Handler`, so the same call wraps a standard-library mux, chi, or gorilla/mux. No adapter is required.

Gin and Echo do not take an `http.Handler` as their router. Adapt at the boundary:

```go
// Gin
r.Use(func(c *gin.Context) {
    apervia.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        c.Request = r
        c.Next()
    })).ServeHTTP(c.Writer, c.Request)
})

// Echo
e.Use(echo.WrapMiddleware(func(next http.Handler) http.Handler {
    return apervia.Middleware(next)
}))
```

Fiber is out of scope. `apervia.WithDevelopmentIdentity` is for a process that is not behind the sidecar.

App tests import `github.com/the-curve-consulting/apervia-sdks/go/aperviatest`. Declare `Personas` once, then sign a request in with `LoginAs`. That package's docs show the setup. It sets the real headers, so the request passes through `apervia.Middleware`.

From `go/`, the SDK suite is `go test -race ./...`.
