package micro

import "net/http"

// Context is request scoped application context
// It manages application request flow
type Context struct {
	Request  *http.Request
	Response http.ResponseWriter

	Params Params

	// Meta is a key/value pair exclusively for the context of each request.
	Meta map[string]interface{}
}

func (c *Context) reset() {
	c.Params = c.Params[0:0]
}

// ContentType returns the Content-Type header of the request.
func (c *Context) ContentType() string {
	return filterFlags(c.Request.Header.Get("Content-Type"))
}

// SetContentType sets Content-Type header to response
func (c *Context) SetContentType(value []string) {
	header := c.Response.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}
