package micro

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

// App holds fully working application setup
type App struct {
	Options

	router *Router
	pool   sync.Pool
}

// New returns an App instance with default configuration.
func New() *App {
	return NewWithOptions(NewOptions())
}

// NewWithOptions creates new application instance
// with given Application Options object
func NewWithOptions(opts Options) *App {

	opts = optionsWithDefault(opts)
	r := NewRouter()
	app := &App{
		Options: opts,
		router:  r,
	}
	//context pool allocation
	app.pool.New = func() interface{} {
		return app.allocateContext()
	}

	return app
}

// GET is a shortcut for routea.router.Handle(http.MethodGet, path, handler)
func (a *App) GET(path string, handler HandlerFunc) {
	a.router.Handle(http.MethodGet, path, handler)
}

// HEAD is a shortcut for routea.router.Handle(http.MethodHead, path, handler)
func (a *App) HEAD(path string, handler HandlerFunc) {
	a.router.Handle(http.MethodHead, path, handler)
}

// OPTIONS is a shortcut for routea.router.Handle(http.MethodOptions, path, handler)
func (a *App) OPTIONS(path string, handler HandlerFunc) {
	a.router.Handle(http.MethodOptions, path, handler)
}

// POST is a shortcut for routea.router.Handle(http.MethodPost, path, handler)
func (a *App) POST(path string, handler HandlerFunc) {
	a.router.Handle(http.MethodPost, path, handler)
}

// PUT is a shortcut for routea.router.Handle(http.MethodPut, path, handler)
func (a *App) PUT(path string, handler HandlerFunc) {
	a.router.Handle(http.MethodPut, path, handler)
}

// PATCH is a shortcut for routea.router.Handle(http.MethodPatch, path, handler)
func (a *App) PATCH(path string, handler HandlerFunc) {
	a.router.Handle(http.MethodPatch, path, handler)
}

// DELETE is a shortcut for routea.router.Handle(http.MethodDelete, path, handler)
func (a *App) DELETE(path string, handler HandlerFunc) {
	a.router.Handle(http.MethodDelete, path, handler)
}

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
func (a *App) Any(path string, handler HandlerFunc) {
	a.router.Handle("GET", path, handler)
	a.router.Handle("POST", path, handler)
	a.router.Handle("PUT", path, handler)
	a.router.Handle("PATCH", path, handler)
	a.router.Handle("HEAD", path, handler)
	a.router.Handle("OPTIONS", path, handler)
	a.router.Handle("DELETE", path, handler)
	a.router.Handle("CONNECT", path, handler)
	a.router.Handle("TRACE", path, handler)
}

func (a *App) allocateContext() *Context {
	return &Context{}
}

// ServeHTTP conforms to the http.Handler interface.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get context from pool
	c := a.pool.Get().(*Context)
	// reset context from previous use
	c.reset()
	c.Request = r
	c.Response = w

	// handle the request
	a.dispatchRequest(c)

	// put back context to pool
	a.pool.Put(c)
}

// dispatchRequest finds appropriate route in routing tree and handles routing rules,
// binds params with context and forwards action to execution
func (a *App) dispatchRequest(c *Context) {
	req := c.Request
	path := c.Request.URL.Path
	if root := a.router.trees[req.Method]; root != nil {
		if route, ps, _ := root.getValue(path); route != nil {
			c.Params = ps
			a.handleAction(c, route.Handler)
			return
		}
	}

	http.NotFound(c.Response, req)
}

// handleAction - handles action results and error reporting
func (a *App) handleAction(c *Context, handler HandlerFunc) {
	res, err := handler(c)
	if err != nil {
		a.Logger.Errorf("action execution returned error: %v", err)
	}

	// handle action result from handler
	err = res.Handle(c)
	if err != nil {
		a.Logger.Errorf("action result returned error: %v", err)
	}
}

// Serve the application at the specified address/port and listen for OS
// interrupt and kill signals and will attempt to stop the application
// gracefully.
func (a *App) Serve() error {
	a.Logger.Info(fmt.Sprintf("Starting Application at %s", a.Addr))
	// create http server
	srv := http.Server{
		Handler: a,
	}

	// make interrupt channel
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, os.Interrupt)
	// listen for interrupt signal
	go func() {
		<-c
		a.Logger.Info("Shutting down application")
		if err := a.stop(); err != nil {
			a.Logger.Error(err.Error())
		}

		if err := srv.Shutdown(context.Background()); err != nil {
			a.Logger.Error(err.Error())
		}
	}()

	if strings.HasPrefix(a.Addr, "unix:") {
		// create unix network listener
		lis, err := net.Listen("unix", a.Addr[5:])
		if err != nil {
			return err
		}
		// start accepting incomming requests on listener
		return srv.Serve(lis)
	}

	//assign address to server
	srv.Addr = a.Addr
	// start accepting incomming requests on listener
	return srv.ListenAndServe()
}

func (a *App) stop() error {
	return nil
}

// Stop issues interrupt signal
func (a *App) Stop() error {
	// get current process
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}
	a.Logger.Debug("Stopping....")
	// issue interrupt signal
	return proc.Signal(os.Interrupt)
}
