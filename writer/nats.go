package writer

import (
	"github.com/skilld-labs/http-event-adapter/log"

	nats "github.com/nats-io/nats.go"
)

type natsWriter struct {
	logger log.Logger
	nats   *nats.Conn
}

var w *natsWriter

func NewNatsWriter(cfg *WriterConfiguration) (Writer, error) {
	if w != nil {
		return w, nil
	}
	nats, err := nats.Connect(cfg.Config.GetString("nats.url"))
	if err != nil {
		return nil, err
	}
	w = &natsWriter{
		logger: cfg.Logger,
		nats:   nats,
	}
	return w, nil
}

func (w *natsWriter) Write(channel string, data []byte) error {
	if err := w.nats.Publish(channel, data); err != nil {
		return err
	}
	w.logger.Debug("an event has been sent in nats (subject: %s)", channel)
	return nil
}
