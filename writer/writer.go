package writer

import (
	"fmt"

	"github.com/skilld-labs/http-event-adapter/configuration"
	"github.com/skilld-labs/http-event-adapter/log"
)

type WriterConfiguration struct {
	Logger log.Logger
	Config configuration.Provider
}

type Writer interface {
	Write(channel string, data []byte) error
}

func GetWriter(cfg *WriterConfiguration, name string) (Writer, error) {
	var writer Writer
	var err error
	switch name {
	case "nats":
		writer, err = NewNatsWriter(cfg)
	default:
		err = fmt.Errorf("unknown writer name %s", name)
	}
	return writer, err
}
