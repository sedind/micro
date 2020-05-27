package micro

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/sedind/micro/render"
)

// ActionResult defines standardized ways of handling HTTP Action Results
type ActionResult interface {
	Handle(c *Context) error
}

type errorResult struct {
	err  error
	code int
}

func (er *errorResult) Handle(c *Context) error {
	var res ActionResult
	switch c.ContentType() {
	case MIMEJSON:
		res = RenderJSON(er.code, VM{"error": er.err.Error()})
	case MIMEXML, MIMEXML2:
		type Error struct {
			Message string
		}
		res = RenderXML(er.code, &Error{Message: er.err.Error()})
	default:
		res = RenderText(er.code, er.err.Error())
	}

	return res.Handle(c)
}

// ErrorResult creates error ActionResult implementation
func ErrorResult(code int, err error) ActionResult {
	return &errorResult{
		code: code,
		err:  err,
	}
}

type redirectResult struct {
	url  string
	code int
}

// Handle finalizes redirect result
func (rr *redirectResult) Handle(c *Context) error {
	http.Redirect(c.Response, c.Request, rr.url, rr.code)
	return nil
}

// RedirectResult creates redirect ActionResult implementation
func RedirectResult(url string, code int) ActionResult {
	return &redirectResult{
		url:  url,
		code: code,
	}
}

type fileResult struct {
	name   string
	reader io.Reader
}

func (fr *fileResult) Handle(c *Context) error {

	h := c.Response.Header()

	ext := filepath.Ext(fr.name)
	t := mime.TypeByExtension(ext)
	if t == "" {
		t = "application/octet-stream"
	}

	cd := fmt.Sprintf("attachment; filename=%s", fr.name)
	//cl := strconv.Itoa(int(written))
	h.Add("Content-Disposition", cd)
	//h.Add("Content-Length", cl)
	h.Add("Content-Type", t)

	_, err := io.Copy(c.Response, fr.reader)
	if err != nil {
		return err
	}

	return err
}

// FileResult creates file attachment ActionResult with following headers:
//
//   Content-Type
//   Content-Length
//   Content-Disposition
//
// Content-Type is set using mime#TypeByExtension with the filename's extension. Content-Type will default to
// application/octet-stream if using a filename with an unknown extension.
func FileResult(name string, reader io.Reader) ActionResult {
	return &fileResult{
		name:   name,
		reader: reader,
	}
}

// RenderResult handlers different rendering implementations
type renderResult struct {
	render.Renderer
	code int
}

// Handle finalizes Render result
func (rr *renderResult) Handle(c *Context) error {
	var res bytes.Buffer
	if err := rr.Render(&res); err != nil {
		return err
	}

	c.SetContentType(rr.ContentType())
	c.Response.WriteHeader(rr.code)

	_, err := c.Response.Write(res.Bytes())

	return err
}

// RenderJSON creates JSON rendered ActionResult
func RenderJSON(code int, data interface{}) ActionResult {
	return &renderResult{
		Renderer: render.JSON{Data: data},
		code:     code,
	}
}

// RenderText creates Text rendered ActionResult
func RenderText(code int, text string) ActionResult {
	return &renderResult{
		Renderer: render.Text{Data: text},
		code:     code,
	}
}

// RenderXML creates XML rendered ActionResult
func RenderXML(code int, data interface{}) ActionResult {
	return &renderResult{
		Renderer: render.XML{Data: data},
		code:     code,
	}
}

//RenderData creates []byte render ActionResult
func RenderData(code int, data []byte, contentType []string) ActionResult {
	return &renderResult{
		Renderer: render.Data{
			Data:  data,
			CType: contentType,
		},
		code: code,
	}
}

// RenderReader creates io.Reader render ActionResult
func RenderReader(code int, reader io.Reader, contentType []string) ActionResult {
	return &renderResult{
		Renderer: render.Reader{
			Reader: reader,
			CType:  contentType,
		},
		code: code,
	}
}
