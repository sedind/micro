package micro

import (
	"bytes"

	"github.com/sedind/micro/render"
)

// ActionResult defines standardized ways of handling HTTP Action Results
type ActionResult interface {
	Handle(c *Context) error
}

// RenderResult handlers different rendering implementations
type RenderResult struct {
	renderer render.Renderer
}

// Handle finalizes Render result
func (rr *RenderResult) Handle(c *Context) error {
	var res bytes.Buffer
	if err := rr.renderer.Render(&res); err != nil {
		return err
	}

	c.SetContentType(rr.renderer.ContentType())

	_, err := c.Response.Write(res.Bytes())

	return err
}

// RenderJSON creates JSON rendered ActionResult
func RenderJSON(data interface{}) ActionResult {
	return &RenderResult{
		renderer: render.JSON{Data: data},
	}
}
