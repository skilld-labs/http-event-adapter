package format

import (
	"fmt"

	"github.com/skilld-labs/http-event-adapter/configuration"
	"github.com/skilld-labs/http-event-adapter/log"
)

type FormatterConfiguration struct {
	Logger log.Logger
	Config configuration.Provider
}

type Formatter interface {
	FormatSingle([]byte) (map[string]interface{}, error)
	FormatMultiple([]byte) ([]interface{}, error)
}

func GetFormatter(cfg *FormatterConfiguration, name string) (Formatter, error) {
	var formatter Formatter
	var err error
	switch name {
	case "json":
		formatter, err = NewJsonFormatter(cfg)
	case "yaml":
		formatter, err = NewYamlFormatter(cfg)
	case "csv":
		formatter, err = NewCsvFormatter(cfg)
	default:
		err = fmt.Errorf("unknown inputFormat %s", name)
	}
	return formatter, err
}
