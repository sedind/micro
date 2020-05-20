package micro

// Route represents a request route's specification which
// contains method and path and its handler.
type Route struct {
	Method string
	Path   string

	Handler HandlerFunc
}
