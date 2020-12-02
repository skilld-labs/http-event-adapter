package format

import "gopkg.in/yaml.v2"

type yamlFormatter struct{}

func NewYamlFormatter(cfg *FormatterConfiguration) (Formatter, error) {
	return &yamlFormatter{}, nil
}

func (y *yamlFormatter) FormatSingle(data []byte) (map[string]interface{}, error) {
	var d map[string]interface{}
	return d, yaml.Unmarshal(data, &d)
}

func (y *yamlFormatter) FormatMultiple(data []byte) ([]interface{}, error) {
	var d []interface{}
	return d, yaml.Unmarshal(data, &d)
}
