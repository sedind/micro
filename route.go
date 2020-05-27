package micro

import "net/http"

// Route represents a request route's specification which
// contains method and path and its handler.
type Route struct {
	Method string
	Path   string

	Mws     *MiddlewareStack
	Handler HandlerFunc
}

// HandleRequest handles user request
func (r *Route) HandleRequest(c *Context) ActionResult {
	if err := r.Mws.handle(c); err != nil {
		return ErrorResult(http.StatusInternalServerError, err)
	}
	return r.Handler(c)
}
