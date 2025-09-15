package lightmux

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func BenchmarkSimpleHandler(b *testing.B) {
	mux := NewLightMux(&http.Server{})
	route := mux.NewRoute("/bench")
	route.Handle(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.ApplyRoutes()
	mux.ApplyGlobalMiddlewares()

	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mux.Mux().ServeHTTP(w, req)
		w.Body.Reset()
	}
}

func BenchmarkWith2MiddlewareHandler(b *testing.B) {
	mux := NewLightMux(&http.Server{})
	route := mux.NewRoute("/bench", func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Add value to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "benchKey", "benchValue")
			r = r.WithContext(ctx)
			next(w, r)
		}
	}, func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				return
			}
		}
	})
	route.Handle(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		// Retrieve value from context
		val := r.Context().Value("benchKey")
		if valStr, ok := val.(string); ok && valStr == "benchValue" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("context missing"))
		}
	})
	mux.ApplyRoutes()
	mux.ApplyGlobalMiddlewares()

	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mux.Mux().ServeHTTP(w, req)
		w.Body.Reset()
	}
}

func BenchmarkWith1MiddlewareHandler(b *testing.B) {
	mux := NewLightMux(&http.Server{})
	route := mux.NewRoute("/bench", func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Add value to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "benchKey", "benchValue")
			r = r.WithContext(ctx)
			next(w, r)
		}
	})
	route.Handle(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		// Retrieve value from context
		val := r.Context().Value("benchKey")
		if valStr, ok := val.(string); ok && valStr == "benchValue" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("context missing"))
		}
	})
	mux.ApplyRoutes()
	mux.ApplyGlobalMiddlewares()

	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mux.Mux().ServeHTTP(w, req)
		w.Body.Reset()
	}
}

func BenchmarkWithLoadedHandlerMany(b *testing.B) {
	mux := NewLightMux(&http.Server{})
	// Register 10,000 dummy routes
	for i := 0; i < 10000; i++ {
		route := mux.NewRoute("/bench" + strconv.Itoa(i))
		route.Handle(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})
	}
	// Register the actual route to benchmark
	route := mux.NewRoute("/bench")
	route.Handle(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.ApplyRoutes()
	mux.ApplyGlobalMiddlewares()

	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mux.Mux().ServeHTTP(w, req)
		w.Body.Reset()
	}
}
