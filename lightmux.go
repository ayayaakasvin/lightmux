// Package lightmux provides a lightweight HTTP server multiplexer with
// support for routing and middleware.
package lightmux

import (
	"context"
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
	server *http.Server   // Should be passed through constructor
	mux    *http.ServeMux // ServeMux that will serve as holder for handlers
}

// NewLightMux creates and returns a new LightMux instance using the provided http.Server.
func NewLightMux(server *http.Server) *LightMux {
	return &LightMux{
		server: server,
		mux:    http.NewServeMux(),
	}
}

// Mux returns the internal http.ServeMux used by LightMux for handler registration.
// This allows direct access to the underlying ServeMux for advanced routing or customization(e.g: adding custom 404 handler).
func (l *LightMux) Mux() *http.ServeMux {
	return l.mux
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
		if err := l.server.ListenAndServe(); err != nil {
			log.Fatalf("ListenAndServe error: %s\n", err)
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
