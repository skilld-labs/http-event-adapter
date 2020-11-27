package format

import "encoding/json"

type jsonFormatter struct{}

func NewJsonFormatter(cfg *FormatterConfiguration) (Formatter, error) {
	return &jsonFormatter{}, nil
}

func (j *jsonFormatter) Format(data []byte) (interface{}, error) {
	var d interface{}
	return d, json.Unmarshal(data, &d)
}
