package micro

import (
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sedind/micro/binding"
	"github.com/sedind/micro/log"
)

var (
	// MaxMultipartMemory used for multipart binding
	MaxMultipartMemory = int64(32 << 20) // 32 MB
)

// Context is request scoped application context
// It manages application request flow
type Context struct {
	Request  *http.Request
	Response http.ResponseWriter

	Params Params

	Logger log.Logger

	// Meta is a key/value pair exclusively for the context of each request.
	Meta map[string]interface{}
}

func (c *Context) reset() {
	c.Params = c.Params[0:0]
}

/************************************/
/*********  APP MANAGEMENT  *********/
/************************************/

// LogFields adds adds structured context to context Logger
//
// This allows you to easily add things
// like metrics (think DB times) to your request.
func (c *Context) LogFields(fields log.Fields) {
	c.Logger = c.Logger.WithFields(fields)
}

/************************************/
/******** METADATA MANAGEMENT********/
/************************************/

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes  c.Meta if it was not used previously.
func (c *Context) Set(key string, value interface{}) {
	if c.Meta == nil {
		c.Meta = make(map[string]interface{})
	}
	c.Meta[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (c *Context) Get(key string) (value interface{}, exists bool) {
	value, exists = c.Meta[key]
	return
}

// GetString returns the value associated with the key as a string.
func (c *Context) GetString(key string) (s string) {
	if val, ok := c.Get(key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}

// GetBool returns the value associated with the key as a boolean.
func (c *Context) GetBool(key string) (b bool) {
	if val, ok := c.Get(key); ok && val != nil {
		b, _ = val.(bool)
	}
	return
}

// GetInt returns the value associated with the key as an integer.
func (c *Context) GetInt(key string) (i int) {
	if val, ok := c.Get(key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// GetInt64 returns the value associated with the key as an integer.
func (c *Context) GetInt64(key string) (i64 int64) {
	if val, ok := c.Get(key); ok && val != nil {
		i64, _ = val.(int64)
	}
	return
}

// GetFloat64 returns the value associated with the key as a float64.
func (c *Context) GetFloat64(key string) (f64 float64) {
	if val, ok := c.Get(key); ok && val != nil {
		f64, _ = val.(float64)
	}
	return
}

// GetTime returns the value associated with the key as time.
func (c *Context) GetTime(key string) (t time.Time) {
	if val, ok := c.Get(key); ok && val != nil {
		t, _ = val.(time.Time)
	}
	return
}

// GetDuration returns the value associated with the key as a duration.
func (c *Context) GetDuration(key string) (d time.Duration) {
	if val, ok := c.Get(key); ok && val != nil {
		d, _ = val.(time.Duration)
	}
	return
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (c *Context) GetStringSlice(key string) (ss []string) {
	if val, ok := c.Get(key); ok && val != nil {
		ss, _ = val.([]string)
	}
	return
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func (c *Context) GetStringMap(key string) (sm map[string]interface{}) {
	if val, ok := c.Get(key); ok && val != nil {
		sm, _ = val.(map[string]interface{})
	}
	return
}

// GetStringMapString returns the value associated with the key as a map of strings.
func (c *Context) GetStringMapString(key string) (sms map[string]string) {
	if val, ok := c.Get(key); ok && val != nil {
		sms, _ = val.(map[string]string)
	}
	return
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func (c *Context) GetStringMapStringSlice(key string) (smss map[string][]string) {
	if val, ok := c.Get(key); ok && val != nil {
		smss, _ = val.(map[string][]string)
	}
	return
}

/************************************/
/************ INPUT DATA ************/
/************************************/

// Param returns the value of the URL param.
//
// It is a shortcut for c.Params.ByName(key)
func (c *Context) Param(key string) string {
	return c.Params.ByName(key)
}

// QueryDefault returns the keyed url query value if it exists,
// otherwise it returns the specified defaultValue string.
//
// See: Query() and GetQuery() for further information.
func (c *Context) QueryDefault(key, defaultValue string) string {
	if value, ok := c.Query(key); ok {
		return value
	}
	return defaultValue
}

// Query returns the keyed url query value
// if it exists `(value, true)` (even when the value is an empty string),
// otherwise it returns `("", false)`.
//
// It is shortcut for `c.Request.URL.Query().Get(key)`
func (c *Context) Query(key string) (string, bool) {
	if values, ok := c.QueryArray(key); ok {
		return values[0], ok
	}
	return "", false
}

// QueryArray returns a slice of strings for a given query key, plus
// a boolean value whether at least one value exists for the given key.
func (c *Context) QueryArray(key string) ([]string, bool) {
	if values, ok := c.Request.URL.Query()[key]; ok && len(values) > 0 {
		return values, true
	}
	return []string{}, false
}

// QueryMap returns a map for a given query key, plus a boolean value
// whether at least one value exists for the given key.
func (c *Context) QueryMap(key string) (map[string]string, bool) {
	return c.get(c.Request.URL.Query(), key)
}

// DefaultPostForm returns the specified key from a POST urlencoded form or multipart form
// when it exists, otherwise it returns the specified defaultValue string.
// See: PostForm() and GetPostForm() for further information.
func (c *Context) DefaultPostForm(key, defaultValue string) string {
	if value, ok := c.PostForm(key); ok {
		return value
	}
	return defaultValue
}

// PostForm returns the specified key from a POST urlencoded
// form or multipart form when it exists `(value, true)` (even when the value is an empty string),
// otherwise it returns ("", false).
func (c *Context) PostForm(key string) (string, bool) {
	if values, ok := c.PostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

// PostFormArray returns a slice of strings for a given form key, plus
// a boolean value whether at least one value exists for the given key.
func (c *Context) PostFormArray(key string) ([]string, bool) {
	req := c.Request
	_ = req.ParseMultipartForm(MaxMultipartMemory)

	if values := req.PostForm[key]; len(values) > 0 {
		return values, true
	}
	if req.MultipartForm != nil && req.MultipartForm.File != nil {
		if values := req.MultipartForm.Value[key]; len(values) > 0 {
			return values, true
		}
	}
	return []string{}, false
}

// GetPostFormMap returns a map for a given form key, plus a boolean value
// whether at least one value exists for the given key.
func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	req := c.Request
	err := req.ParseMultipartForm(MaxMultipartMemory)
	if err != nil {
		panic(err)
	}

	dicts, exist := c.get(req.PostForm, key)

	if !exist && req.MultipartForm != nil && req.MultipartForm.File != nil {
		dicts, exist = c.get(req.MultipartForm.Value, key)
	}

	return dicts, exist
}

// get is an internal method and returns a map which satisfy conditions.
func (c *Context) get(m map[string][]string, key string) (map[string]string, bool) {
	dicts := make(map[string]string)
	exist := false
	for k, v := range m {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dicts[k[i+1:][:j]] = v[0]
			}
		}
	}
	return dicts, exist
}

// FormFile returns the first file for the provided form key.
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if c.Request.MultipartForm == nil {
		if err := c.Request.ParseMultipartForm(MaxMultipartMemory); err != nil {
			return nil, err
		}
	}
	_, fh, err := c.Request.FormFile(name)
	return fh, err
}

// MultipartForm is the parsed multipart form, including file uploads.
func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.Request.ParseMultipartForm(MaxMultipartMemory)
	return c.Request.MultipartForm, err
}

// SaveUploadedFile uploads the form file to specific dest.
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dest string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	io.Copy(out, src)
	return nil
}

/************************************/
/************* BINDING **************/
/************************************/

// Bind checks the Content-Type to select a binding engine automatically,
func (c *Context) Bind(obj interface{}) error {
	b := binding.Default(c.Request.Method, c.ContentType())
	return c.BindWith(obj, b)
}

// BindJSON binds the passed struct pointer using JSON binding engine.
func (c *Context) BindJSON(obj interface{}) error {
	return c.BindWith(obj, binding.JSON)
}

// BindXML binds the passed struct pointer using XML binding engine.
func (c *Context) BindXML(obj interface{}) error {
	return c.BindWith(obj, binding.XML)
}

// BindQuery binds the passed struct pointer using Query binding engine.
func (c *Context) BindQuery(obj interface{}) error {
	return c.BindWith(obj, binding.Query)
}

// BindURI binds the passed struct pointer using URI binding engine.
func (c *Context) BindURI(obj interface{}) error {
	m := make(map[string][]string)
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	return binding.URI.BindURI(m, obj)
}

// BindWith binds the passed struct pointer using the specified binding engine.
// See the binding package.
func (c *Context) BindWith(obj interface{}, b binding.Binder) error {
	return b.Bind(c.Request, obj)
}

// ClientIP implements a best effort algorithm to return the real client IP
//
// it parses X-Real-IP and X-Forwarded-For in order to work properly
// with reverse-proxies such us: nginx or haproxy.
// Use X-Forwarded-For before X-Real-Ip as nginx uses X-Real-Ip with the proxy's IP.
func (c *Context) ClientIP() string {

	clientIP := c.Request.Header.Get("X-Forwarded-For")
	clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	if clientIP == "" {
		clientIP = strings.TrimSpace(c.Request.Header.Get("X-Real-Ip"))
	}
	if clientIP != "" {
		return clientIP
	}

	if addr := c.Request.Header.Get("X-Appengine-Remote-Addr"); addr != "" {
		return addr
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

// RequestID implements a best effort algorithm to return tracing request ID for current request
//
// it parses X-Request-ID which is ment to be an application level tracing id
// and X-Amzn-Trace-Id which is automatically added by Amazon loadbalancers
func (c *Context) RequestID() string {
	// check if request ID exists in headers
	requestID := c.Request.Header.Get("X-Request-ID")

	if requestID == "" {
		//check if  X-Amzn-Trace-Id exists
		requestID = c.Request.Header.Get("X-Amzn-Trace-Id")
	}
	return requestID
}

// ContentType returns the Content-Type header of the request.
func (c *Context) ContentType() string {
	ct := c.Request.Header.Get("Content-Type")
	for i, char := range ct {
		if char == ' ' || char == ';' {
			return ct[:i]
		}
	}
	return ct
}

// SetContentType sets Content-Type header to response
func (c *Context) SetContentType(value []string) {
	header := c.Response.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}

// SetHeader is a intelligent shortcut for c.Response.Header().Set(key, value).
// It writes a header in the response.
// If value == "", this method removes the header `c.Response.Header().Del(key)`
func (c *Context) SetHeader(key, value string) {
	if value == "" {
		c.Response.Header().Del(key)
		return
	}
	c.Response.Header().Set(key, value)
}

// RequestHeader returns value from request header for given key
func (c *Context) RequestHeader(key string) string {
	return c.Request.Header.Get(key)
}

// RequestHeaders returns value from request headers
func (c *Context) RequestHeaders() http.Header {
	return c.Request.Header
}

// ResponseHeader returns value from response header for given key
func (c *Context) ResponseHeader(key string) string {
	return c.Response.Header().Get(key)
}

// ResponseHeaders returns value from response headers
func (c *Context) ResponseHeaders() http.Header {
	return c.Request.Header
}

// RequestBody return stream data from request Body
func (c *Context) RequestBody() ([]byte, error) {
	return ioutil.ReadAll(c.Request.Body)
}

// SetCookie adds a Set-Cookie header to the ResponseWriter's headers.
// The provided cookie must have a valid Name. Invalid cookies may be
// silently dropped.
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.Response, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

// Cookie returns the named cookie provided in the request or
// ErrNoCookie if not found. And return the named cookie is unescaped.
// If multiple cookies match the given name, only one cookie will
// be returned.
func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	val, _ := url.QueryUnescape(cookie.Value)
	return val, nil
}

// Deadline returns the time when work done on behalf of this context
// should be canceled.
//
// Deadline returns ok==false when no deadline is
// set. Successive calls to Deadline return the same results.
func (c *Context) Deadline() (time.Time, bool) {
	return c.Request.Context().Deadline()
}

// Done returns a channel that's closed when work done on behalf of this
// context should be canceled.
//
// Done may return nil if this context can
// never be canceled. Successive calls to Done return the same value.
func (c *Context) Done() <-chan struct{} {
	return c.Request.Context().Done()
}

// Err returns a non-nil error value after Done is closed,
// successive calls to Err return the same error.
//
// If Done is not yet closed, Err returns nil.
// If Done is closed, Err returns a non-nil error explaining why:
// Canceled if the context was canceled
// or DeadlineExceeded if the context's deadline passed.
func (c *Context) Err() error {
	return c.Request.Context().Err()
}
