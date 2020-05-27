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
		res = JSONResult(er.code, VM{"error": er.err.Error()})
	case MIMEYAML:
		res = YAMLResult(er.code, VM{"error": er.err.Error()})
	case MIMEXML, MIMEXML2:
		type Error struct {
			Message string
		}
		res = XMLResult(er.code, &Error{Message: er.err.Error()})
	default:
		res = TextResult(er.code, er.err.Error())
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
	filepath string
}

func (rf *fileResult) Handle(c *Context) error {
	http.ServeFile(c.Response, c.Request, rf.filepath)
	return nil
}

// FileResult writes the specified file into the body stream in a efficient way.
func FileResult(filePath string) ActionResult {
	return &fileResult{
		filepath: filePath,
	}
}

type downloadResult struct {
	name   string
	reader io.Reader
}

func (fr *downloadResult) Handle(c *Context) error {

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

// DownloadResult creates file attachment ActionResult with following headers:
//
//   Content-Type
//   Content-Length
//   Content-Disposition
//
// Content-Type is set using mime#TypeByExtension with the filename's extension. Content-Type will default to
// application/octet-stream if using a filename with an unknown extension.
func DownloadResult(name string, reader io.Reader) ActionResult {
	return &downloadResult{
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

// JSONResult creates JSON rendered ActionResult
func JSONResult(code int, data interface{}) ActionResult {
	return &renderResult{
		Renderer: render.JSON{Data: data},
		code:     code,
	}
}

// YAMLResult creates YAML rendered ActionResult
func YAMLResult(code int, data interface{}) ActionResult {
	return &renderResult{
		Renderer: render.YAML{Data: data},
		code:     code,
	}
}

// TextResult creates Text rendered ActionResult
func TextResult(code int, text string) ActionResult {
	return &renderResult{
		Renderer: render.Text{Data: text},
		code:     code,
	}
}

// XMLResult creates XML rendered ActionResult
func XMLResult(code int, data interface{}) ActionResult {
	return &renderResult{
		Renderer: render.XML{Data: data},
		code:     code,
	}
}

//DataResult creates []byte render ActionResult
func DataResult(code int, data []byte, contentType []string) ActionResult {
	return &renderResult{
		Renderer: render.Data{
			Data:  data,
			CType: contentType,
		},
		code: code,
	}
}

// ReaderResult creates io.Reader render ActionResult
func ReaderResult(code int, reader io.Reader, contentType []string) ActionResult {
	return &renderResult{
		Renderer: render.Reader{
			Reader: reader,
			CType:  contentType,
		},
		code: code,
	}
}
