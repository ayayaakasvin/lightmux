package lightmux

import (
	"fmt"
	"net/http"
)

var middlewareStack []GlobalMiddlewareFunc

// GlobalMiddlewareFunc defines a function type for global HTTP middleware.
type GlobalMiddlewareFunc func(http.Handler) http.Handler

// Global middleware functions are applied to all incoming HTTP requests handled by the server.
// Func registers GlobalMiddlewareFunc and can be used for logging, authentication, etc.
// Changes will be applied to server after runnung LightMux.Run func.
func (l *LightMux) Use(middlewares ...GlobalMiddlewareFunc) {
	if len(middlewares) != 0 {
		middlewareStack = append(middlewareStack, middlewares...)
	}
}

func chainMiddlewares(handler http.Handler, middlewares []GlobalMiddlewareFunc) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// ApplyGlobalMiddlewares applies all registered global middlewares to the HTTP handler.
// This method is called after all routes have been registered and
// before starting the HTTP server (inside Run() method).
func (l *LightMux) ApplyGlobalMiddlewares() {
	finalHandler := http.Handler(l.mux)
	if len(middlewareStack) > 0 {
		finalHandler = chainMiddlewares(finalHandler, middlewareStack)
	}
	l.server.Handler = finalHandler
}

// Prints count of registered middlewares
func (l *LightMux) PrintMiddlewareInfo() {
	fmt.Printf("Global middleware count: %d\n", len(middlewareStack))
}
