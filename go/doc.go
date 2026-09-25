// Package apervia reads the platform identity from a request.
//
// Middleware stores an Identity on the request context. FromContext returns it.
// An anonymous request has a false second value. Identity carries app-plane
// permissions and role names only.
package apervia
