package log

import "io"

// Configuration stores the config for the logger
// For some loggers there can only be one level across writers, for such the level of Console is picked by default
type Configuration struct {
	JSONFormat bool
	Level      string
	Output     io.Writer
}
