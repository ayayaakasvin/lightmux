package lightmux

// RouteGroup represents a group of routes with a common prefix and shared middlewares.
type RouteGroup struct {
	prefix      string
	middlewares []Middleware
	mux         *LightMux
}

// NewGroup creates a new RouteGroup with the given prefix and optional middlewares.
func (l *LightMux) NewGroup(prefix string, middlewares ...Middleware) *RouteGroup {
	return &RouteGroup{
		prefix:      prefix,
		middlewares: middlewares,
		mux:         l,
	}
}

// NewRoute creates a new Route within the RouteGroup with the given path and optional middlewares.
func (g *RouteGroup) NewRoute(path string, middlewares ...Middleware) *Route {
	fullPath := g.prefix + path
	allMiddleware := append(g.middlewares, middlewares...)
	return g.mux.NewRoute(fullPath, allMiddleware...)
}

// 
func (g *RouteGroup) ContinueGroup(path string, middlewares ...Middleware) *RouteGroup {
	newPrefix := g.prefix + path
	
	newMiddlewares := make([]Middleware, len(g.middlewares))
	copy(newMiddlewares, g.middlewares)

	newMiddlewares = append(newMiddlewares, middlewares...)

	newGroup := &RouteGroup{
		prefix: newPrefix,
		middlewares: newMiddlewares,
		mux: g.mux,
	}

	return newGroup
}