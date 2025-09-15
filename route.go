package lightmux

import (
	"fmt"
	"net/http"
)

// Middleware defines a function to process middleware logic for HTTP handlers.
type Middleware func(http.HandlerFunc) http.HandlerFunc

// Route represents an HTTP route with its path, supported methods, and middlewares.
type Route struct {
	Path        	string
	Methods     	map[string]http.Handler
	Middlewares 	[]Middleware
}

// NewRoute creates a new Route with the given path and optional middlewares.
func (l *LightMux) NewRoute(path string, middlewares ...Middleware) *Route {
	// Check for duplicate path
	if _, exists := l.routeMap[path]; exists {
		panic(fmt.Sprintf("route with path %v already exists", path))
	}

	r := &Route{
		Path:        path,
		Methods:     make(map[string]http.Handler),
		Middlewares: middlewares,
	}

	l.routeMap[path] = r

	return r
}

// Use adds middlewares into route middlewares.
func (r *Route) Use(middlewares ...Middleware)  {
	r.Middlewares = append(r.Middlewares, middlewares...)
}

// Handle registers a handler for a specific HTTP method on the route.
// Middlewares are not wrapped here; they are applied when serving the request.
func (r *Route) Handle(method string, handler http.HandlerFunc) {
	if !isValidMethod(method) {
		panic("invalid HTTP method: " + method)
	}

	// check if method already exists
	if _, exists := r.Methods[method]; exists {
		panic("duplicate method for path: " + method + " " + r.Path)
	}

	r.Methods[method] = r.wrapMiddlewares(handler)
}

// wrapMiddlewares applies the route's middlewares to the given handler.
func (r *Route) wrapMiddlewares(handler http.HandlerFunc) http.HandlerFunc {
	for i := len(r.Middlewares) - 1; i >= 0; i-- {
		handler = r.Middlewares[i](handler)
	}
	return handler
}