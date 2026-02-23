// Package lightmux provides a lightweight HTTP server multiplexer with
// support for routing and middleware.
package lightmux

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// LightMux is the main struct that manages the HTTP server and routing.
// It holds a reference to an http.Server and an http.ServeMux for handler registration.
type LightMux struct {
	server *http.Server   // HTTP server instance managed by LightMux.
	mux    *http.ServeMux // ServeMux that will serve as holder for handlers.

	// routeMap is a map for quick lookup of registered route patterns.
	routeMap map[string]*Route

	// globalMiddlewareStack holds the stack of global middlewares applied to all routes.
	globalMiddlewareStack []Middleware
}

// NewLightMux creates and returns a new LightMux instance using the provided http.Server.
func NewLightMux(server *http.Server) *LightMux {
	return &LightMux{
		server:   server,
		mux:      http.NewServeMux(),
		routeMap: make(map[string]*Route),
	}
}

// Mux returns the internal http.ServeMux used by LightMux for handler registration.
// This allows direct access to the underlying ServeMux for advanced routing or customization(e.g: adding custom 404 handler).
func (l *LightMux) Mux() *http.ServeMux {
	return l.mux
}

// ApplyRoutes registers all routes that have been created with NewRoute.
//
// Run() calls this before starting HTTP server, and before applying any global middlewares.
// This ensures all route handlers are registered to the underlying mux.
func (l *LightMux) ApplyRoutes() {
	for _, route := range l.routeMap {
		route := route
		allowed := allowedMethodsJoin(route.Methods)

		l.mux.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			if handler, ok := route.Methods[r.Method]; ok {
				handler.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"error": fmt.Sprintf("%s method is not allowed, allowed methods for %s:[%s]", r.Method, r.URL.Path, allowed),
				})
				return
			}
		})
	}
}

// PrintRoutes prints all registered routes and their supported methods.
func (l *LightMux) PrintRoutes() {
	for _, r := range l.routeMap {
		fmt.Printf("Route: %s\n", r.Path)
		for method, handler := range r.Methods {
			fmt.Printf("\t- %s (handler: %s)\n", method, getFuncName(handler))
		}
		fmt.Printf("\tMiddlewares: %d\n", len(r.Middlewares))
		for i, mw := range r.Middlewares {
			fmt.Printf("\t\t%d: %T (%s)\n", i+1, mw, getFuncName(mw))
		}
	}
}

// Run starts the HTTP server and blocks until the server stops.
// It returns any error encountered while running the server.
// The caller is responsible for managing context cancellation and graceful shutdown.
func (l *LightMux) Run(ctx context.Context) error {
	l.ApplyRoutes()
	l.ApplyGlobalMiddlewares()

	errCh := make(chan error, 1)

	go func() {
		log.Println("Starting LightMux on", l.server.Addr)
		if err := l.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := l.server.Shutdown(shutdownCtx); err != nil {
			return err
		}

		log.Println("Server shutdown complete.")
		return nil

	case err := <-errCh:
		return err
	}
}

// RunTLS starts the HTTP server with TLS support.
// It applies all registered routes and global middlewares before starting the server.
// The caller is responsible for managing context cancellation and graceful shutdown.
// Parameters:
// - ctx: Context for managing server lifecycle.
// - certFile: Path to the TLS certificate file.
// - keyFile: Path to the TLS key file.
// Returns:
// - An error if the server fails to start or shut down properly.
func (l *LightMux) RunTLS(ctx context.Context, certFile, keyFile string) error {
	l.ApplyRoutes()
	l.ApplyGlobalMiddlewares()

	errCh := make(chan error, 1)

	go func() {
		log.Println("Starting LightMux on", l.server.Addr)
		if err := l.server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := l.server.Shutdown(shutdownCtx); err != nil {
			return err
		}

		log.Println("Server shutdown complete.")
		return nil

	case err := <-errCh:
		return err
	}
}
