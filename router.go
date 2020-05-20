// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/julienschmidt/httprouter/blob/master/LICENSE

package micro

import (
	"net/http"
	"strings"
)

// Router is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router struct {
	trees map[string]*node

	mws []HandlerFunc
}

// NewRouter returns a new initialized Router.
// Path auto-correction, including trailing slashes, is enabled by default.
func NewRouter() *Router {
	return &Router{}
}

// GET is a shortcut for router.Handle(http.MethodGet, path, handler)
func (r *Router) GET(path string, handler HandlerFunc) {
	r.Handle(http.MethodGet, path, handler)
}

// HEAD is a shortcut for router.Handle(http.MethodHead, path, handler)
func (r *Router) HEAD(path string, handler HandlerFunc) {
	r.Handle(http.MethodHead, path, handler)
}

// OPTIONS is a shortcut for router.Handle(http.MethodOptions, path, handler)
func (r *Router) OPTIONS(path string, handler HandlerFunc) {
	r.Handle(http.MethodOptions, path, handler)
}

// POST is a shortcut for router.Handle(http.MethodPost, path, handler)
func (r *Router) POST(path string, handler HandlerFunc) {
	r.Handle(http.MethodPost, path, handler)
}

// PUT is a shortcut for router.Handle(http.MethodPut, path, handler)
func (r *Router) PUT(path string, handler HandlerFunc) {
	r.Handle(http.MethodPut, path, handler)
}

// PATCH is a shortcut for router.Handle(http.MethodPatch, path, handler)
func (r *Router) PATCH(path string, handler HandlerFunc) {
	r.Handle(http.MethodPatch, path, handler)
}

// DELETE is a shortcut for router.Handle(http.MethodDelete, path, handler)
func (r *Router) DELETE(path string, handler HandlerFunc) {
	r.Handle(http.MethodDelete, path, handler)
}

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
func (r *Router) Any(path string, handler HandlerFunc) {
	r.Handle("GET", path, handler)
	r.Handle("POST", path, handler)
	r.Handle("PUT", path, handler)
	r.Handle("PATCH", path, handler)
	r.Handle("HEAD", path, handler)
	r.Handle("OPTIONS", path, handler)
	r.Handle("DELETE", path, handler)
	r.Handle("CONNECT", path, handler)
	r.Handle("TRACE", path, handler)
}

// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (r *Router) Handle(method, path string, handler HandlerFunc) {
	//varsCount := uint16(0)

	if method == "" {
		panic("method must not be empty")
	}
	if len(path) < 1 || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}
	if handler == nil {
		panic("handler must not be nil")
	}

	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root
	}

	route := &Route{
		Method:  method,
		Path:    path,
		Handler: handler,
	}

	root.addRoute(path, route)

}

// Lookup allows the manual lookup of a method + path combo.
//
// If the path was found, it returns the handler function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (r *Router) Lookup(method, path string) (*Route, Params, bool) {
	if root := r.trees[method]; root != nil {
		return root.getValue(path)
	}
	return nil, nil, false
}

func (r *Router) allowed(path, reqMethod string) (allow string) {
	allowed := make([]string, 0, 9)

	if path == "*" { // server-wide
		for method := range r.trees {
			if method == http.MethodOptions {
				continue
			}
			// Add request method to list of allowed methods
			allowed = append(allowed, method)
		}
	} else { // specific path
		for method := range r.trees {
			// Skip the requested method - we already tried this one
			if method == reqMethod || method == http.MethodOptions {
				continue
			}

			handle, _, _ := r.trees[method].getValue(path)
			if handle != nil {
				// Add request method to list of allowed methods
				allowed = append(allowed, method)
			}
		}
	}

	if len(allowed) > 0 {
		// Add request method to list of allowed methods
		allowed = append(allowed, http.MethodOptions)

		// Sort allowed methods.
		// sort.Strings(allowed) unfortunately causes unnecessary allocations
		// due to allowed being moved to the heap and interface conversion
		for i, l := 1, len(allowed); i < l; i++ {
			for j := i; j > 0 && allowed[j] < allowed[j-1]; j-- {
				allowed[j], allowed[j-1] = allowed[j-1], allowed[j]
			}
		}

		// return as comma separated list
		return strings.Join(allowed, ", ")
	}
	return
}
