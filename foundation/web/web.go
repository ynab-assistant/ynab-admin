package web

import (
	"context"
	"net/http"
	"os"
	"syscall"

	"github.com/go-chi/chi/v5"
)

// A Handler is a type that handles an http request within our own little mini
// framework.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// registered keeps track of handlers registered to the http default server
// mux. This is a singleton and used by the standard library for metrics
// and profiling. The application may want to add other handlers like
// readiness and liveness to that mux. If this is not tracked, the routes
// could try to be registered more than once, causing a panic.
var registered = make(map[string]bool)

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers.
type App struct {
	mux      *chi.Mux
	shutdown chan os.Signal
	mw       []Middleware
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {

	mux := chi.NewMux()

	return &App{
		mux:      mux,
		shutdown: shutdown,
		mw:       mw,
	}
}

// SignalShutdown is used to gracefully shutdown the app when an integrity
// issue is identified.
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// ServeHTTP implements the http.Handler interface. It's the entry point for
// all http traffic and allows the opentelemetry mux to run first to handle
// tracing. The opentelemetry mux then calls the application mux to handle
// application traffic. This was setup on line 58 in the NewApp function.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

// HandleDebug sets a handler function for a given HTTP method and path pair
// to the default http package server mux. /debug is added to the path.
func (a *App) HandleDebug(method string, path string, handler Handler, mw ...Middleware) {
	a.handle(true, method, path, handler, mw...)
}

// Handle sets a handler function for a given HTTP method and path pair
// to the application server mux.
func (a *App) Handle(method string, path string, handler Handler, mw ...Middleware) {
	a.handle(false, method, path, handler, mw...)
}

// handle performs the real work of applying boilerplate and framework code
// for a handler.
func (a *App) handle(debug bool, method string, path string, handler Handler, mw ...Middleware) {
	if debug {
		// Track all the handlers that are being registered so we don't have
		// the same handlers registered twice to this singleton.
		if _, exists := registered[method+path]; exists {
			return
		}
		registered[method+path] = true
	}

	// First wrap handler specific middleware around this handler.
	handler = wrapMiddleware(mw, handler)

	// Add the application's general middleware to the handler chain.
	handler = wrapMiddleware(a.mw, handler)

	// The function to execute for each request.
	h := func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		// Call the wrapped handler functions.
		if err := handler(ctx, w, r); err != nil {
			a.SignalShutdown()
			return
		}
	}

	// Add this handler for the specified verb and route.
	if debug {
		f := func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == method:
				h(w, r)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}
		http.DefaultServeMux.HandleFunc("/debug"+path, f)
		return
	}
	a.mux.MethodFunc(method, path, h)
}
