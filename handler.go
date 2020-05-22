package micro

// HandlerFunc is a function that can be registered to a route to handle HTTP
// requests.
type HandlerFunc func(*Context) ActionResult
