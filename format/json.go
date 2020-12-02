package format

import "encoding/json"

type jsonFormatter struct{}

func NewJsonFormatter(cfg *FormatterConfiguration) (Formatter, error) {
	return &jsonFormatter{}, nil
}

func (j *jsonFormatter) FormatSingle(data []byte) (map[string]interface{}, error) {
	var d map[string]interface{}
	return d, json.Unmarshal(data, &d)
}

func (j *jsonFormatter) FormatMultiple(data []byte) ([]interface{}, error) {
	var d []interface{}
	return d, json.Unmarshal(data, &d)
}
