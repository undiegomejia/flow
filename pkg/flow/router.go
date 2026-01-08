// Package flow: public router adapter.
//
// This file exposes a framework-friendly Router that wraps the internal
// routing engine. It keeps the public API small: users of the framework
// should import pkg/flow and use flow.NewRouter(app) to register routes
// and resources using idiomatic types (flow.Resource, flow.Controller).
package flow

import (
	"net/http"

	routerpkg "github.com/dministrator/flow/internal/router"
)

// Router is the public wrapper around internal/router.Router. It accepts
// framework Resource implementations and controller handlers and exposes
// a small, testable surface.
type Router struct {
	inner *routerpkg.Router
	app   *App
}

// NewRouter constructs a Router bound to the provided App. App may be nil
// for tests, but Resource adapters that need App will require a non-nil
// App to function correctly.
func NewRouter(app *App) *Router {
	return &Router{inner: routerpkg.New(), app: app}
}

// Get registers a GET handler for the given pattern.
func (r *Router) Get(pattern string, h http.HandlerFunc) { r.inner.Get(pattern, h) }

// Post registers a POST handler for the given pattern.
func (r *Router) Post(pattern string, h http.HandlerFunc) { r.inner.Post(pattern, h) }

// Put registers a PUT handler for the given pattern.
func (r *Router) Put(pattern string, h http.HandlerFunc) { r.inner.Put(pattern, h) }

// Patch registers a PATCH handler for the given pattern.
func (r *Router) Patch(pattern string, h http.HandlerFunc) { r.inner.Patch(pattern, h) }

// Delete registers a DELETE handler for the given pattern.
func (r *Router) Delete(pattern string, h http.HandlerFunc) { r.inner.Delete(pattern, h) }

// Resources wires a flow.Resource into RESTful routes using the conventional
// path base. It uses MakeResourceAdapter to adapt the Resource to the
// internal router.ResourceController.
func (r *Router) Resources(base string, res Resource) error {
	if r.app == nil {
		// adapter requires app for context; return early to avoid panics
		return r.inner.Resources(base, nil)
	}
	return r.inner.Resources(base, MakeResourceAdapter(r.app, res))
}

// ServeHTTP forwards to the internal router's ServeHTTP implementation.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.inner.ServeHTTP(w, req)
}

// Handler returns the underlying http.Handler so the Router can be used
// directly with net/http servers.
func (r *Router) Handler() http.Handler { return r.inner }
