package render

import (
	"io"
)

// Renderer defines for rendering different content types
type Renderer interface {
	// Render writes data to io.Writer
	Render(io.Writer) error

	// ContentType returns contentType for renderer
	ContentType() []string
}
