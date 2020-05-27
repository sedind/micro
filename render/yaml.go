package render

import (
	"io"

	"gopkg.in/yaml.v2"
)

var yamlContentType = []string{"application/x-yaml; charset=utf-8"}

// YAML renders data as YAML content type
type YAML struct {
	Data interface{}
}

// Render YAML content to io.Writer
func (r YAML) Render(out io.Writer) error {
	data, err := yaml.Marshal(r.Data)
	if err != nil {
		return err
	}
	out.Write(data)
	return nil
}

// ContentType returns contentType for renderer
func (YAML) ContentType() []string {
	return yamlContentType
}
