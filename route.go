package lightmux

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// routeStack holds all registered routes.
var (
	routeStack 	[]*Route
	routeMap 	map[string]struct{} = make(map[string]struct{})
)

// Middleware defines a function to process middleware logic for HTTP handlers.
type Middleware func(http.HandlerFunc) http.HandlerFunc

// Route represents an HTTP route with its path, supported methods, and middlewares.
type Route struct {
	Path        string
	Methods     map[string]http.HandlerFunc
	Middlewares []Middleware
}

func isValidMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete,
		http.MethodPatch, http.MethodHead, http.MethodOptions, http.MethodConnect, http.MethodTrace:
		return true
	default:
		return false
	}
}

// NewRoute creates a new Route with the given path and optional middlewares.
func (l *LightMux) NewRoute(path string, middlewares ...Middleware) *Route {
	// Check for duplicate path
	if _, exists := routeMap[path]; exists {
		panic(fmt.Sprintf("route with path %v already exists", path))
	}

	r := &Route{
		Path:        path,
		Methods:     make(map[string]http.HandlerFunc),
		Middlewares: middlewares,
	}

	routeStack = append(routeStack, r)
	routeMap[path] = struct{}{}

	return r
}

// Handle registers a handler for a specific HTTP method on the route.
func (r *Route) Handle(method string, handler http.HandlerFunc) {
	if !isValidMethod(method) {
		panic("invalid HTTP method: " + method)
	}

	// wrap middlewares for handler
	for i := len(r.Middlewares) - 1; i >= 0; i-- {
		handler = r.Middlewares[i](handler)
	}

	// check if method already exists
	if _, exists := r.Methods[method]; exists {
		panic("duplicate method for path: " + method + " " + r.Path)
	}

	r.Methods[method] = handler
}

// ApplyRoutes registers all routes that have been created with NewRoute.
//
// Run() calls this before starting HTTP server, and before applying any global middlewares.
// This ensures all route handlers are registered to the underlying mux.
func (l *LightMux) ApplyRoutes() {
	for _, route := range routeStack {
		route := route
		l.mux.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			if handler, ok := route.Methods[r.Method]; ok {
				handler.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"error": fmt.Sprintf("No handler for method %s for path %s", r.Method, r.URL.Path),
				})
				return
			}
		})
	}
}

// PrintRoutes prints all registered routes and their supported methods.
func (l *LightMux) PrintRoutes() {
	for _, r := range routeStack {
		fmt.Printf("Route: %s\n", r.Path)
		for method := range r.Methods {
			fmt.Printf("\t- %s\n", method)
		}
	}
}
