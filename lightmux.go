// Package lightmux provides a lightweight HTTP server multiplexer with
// support for routing and middleware.
package lightmux

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

// Run applies routes and global middlewares, then starts the HTTP server.
// It returns any error encountered while running the server.
// When server is stopped, it shutdowns gracefully.
func (l *LightMux) Run() error {
	l.ApplyRoutes()
	l.ApplyGlobalMiddlewares()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Starting LightMux on", l.server.Addr)
		if err := l.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %s\n", err)
		} else if err == http.ErrServerClosed {
			log.Println("Server closed gracefully.")
			os.Exit(0)
		}
	}()

	<-stop
	log.Println("Shutdown signal received, shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := l.server.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Shutdown failed: %v", err)
	}

	log.Println("Server shutdown complete.")
	return nil
}

// Same with *LightMux.Run(), but with custom context
func (l *LightMux) RunContext(ctx context.Context) error {
	l.ApplyRoutes()
	l.ApplyGlobalMiddlewares()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Starting LightMux on", l.server.Addr)
		if err := l.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %s\n", err)
		} else if err == http.ErrServerClosed {
			log.Println("Server closed gracefully.")
			os.Exit(0)
		}
	}()

	<-stop
	log.Println("Shutdown signal received, shutting down server...")

	childCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	if err := l.server.Shutdown(childCtx); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Shutdown failed: %v", err)
	}

	log.Println("Server shutdown complete.")
	return nil
}

// RunTLS starts the HTTP server with TLS support using the provided certificate and key files.
// It applies all registered routes and global middlewares before starting the server.
// The server listens for termination signals (e.g., SIGTERM) and shuts down gracefully.
// Parameters:
// - certFile: Path to the TLS certificate file.
// - keyFile: Path to the TLS key file.
// Returns:
// - An error if the server fails to start or shut down properly.
func (l *LightMux) RunTLS(certFile, keyFile string) error {
	l.ApplyRoutes()
	l.ApplyGlobalMiddlewares()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Starting LightMux on", l.server.Addr)
		if err := l.server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServeTLS error: %s\n", err)
		} else if err == http.ErrServerClosed {
			log.Println("Server closed gracefully.")
			os.Exit(0)
		}
	}()

	<-stop
	log.Println("Shutdown signal received, shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := l.server.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Shutdown failed: %v", err)
	}

	log.Println("Server shutdown complete.")
	return nil
}
