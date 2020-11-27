package format

import "gopkg.in/yaml.v2"

type yamlFormatter struct{}

func NewYamlFormatter(cfg *FormatterConfiguration) (Formatter, error) {
	return &yamlFormatter{}, nil
}

func (y *yamlFormatter) Format(data []byte) (interface{}, error) {
	var d interface{}
	return d, yaml.Unmarshal(data, &d)
}
