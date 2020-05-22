package micro

import (
	"os"

	"github.com/sedind/micro/log"
)

const (
	defaultEnv     = "development"
	defaultName    = "FlowApp"
	defaultAddr    = "0.0.0.0:5000"
	defaultVersion = "v0.0.0"

	defaultLogLevel = "debug"

	defaultRedirectTrailingSlash  = true
	defaultRedirectFixedPath      = true
	defaultHandleMethodNotAllowed = true

	default404Body = "404 page not found"
	default405Body = "405 method not allowed"
)

// Options holds flow configuration options
type Options struct {
	Env     string
	Name    string
	Addr    string
	Version string

	LogLevel string
	Logger   log.Logger

	RedirectTrailingSlash  bool
	RedirectFixedPath      bool
	HandleMethodNotAllowed bool

	Body404 string
	Body405 string

	RequestLoggerIgnore []string

	AppConfig interface{}
}

// NewOptions returns a new Options instance with default configuration
func NewOptions() Options {
	opts := Options{
		Env:      defaultEnv,
		Name:     defaultName,
		Version:  defaultVersion,
		Addr:     defaultAddr,
		LogLevel: defaultLogLevel,

		RedirectTrailingSlash:  defaultRedirectTrailingSlash,
		RedirectFixedPath:      defaultRedirectFixedPath,
		HandleMethodNotAllowed: defaultHandleMethodNotAllowed,

		Body404: default404Body,
		Body405: default405Body,
	}

	return opts
}

func optionsWithDefault(opts Options) Options {
	//configure logger
	if opts.Logger == nil {
		opts.Logger = log.New(log.Configuration{
			JSONFormat: true,
			Level:      opts.LogLevel,
			Output:     os.Stdout,
		})
	}

	return opts
}
