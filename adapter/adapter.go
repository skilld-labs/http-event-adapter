package adapter

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"path/filepath"
	"plugin"
	gotemplate "text/template"

	"golang.org/x/sync/errgroup"

	"github.com/skilld-labs/http-event-adapter/log"

	"github.com/skilld-labs/http-event-adapter/configuration"

	"github.com/skilld-labs/http-event-adapter/format"
	"github.com/skilld-labs/http-event-adapter/template"
	"github.com/skilld-labs/http-event-adapter/writer"
)

const (
	defaultBatchInterval = time.Second
)

var (
	errInputInvalid = errors.New("input is invalid")
)

type AdapterConfiguration struct {
	Logger          log.Logger
	Config          configuration.Provider
	FormatterByName func(string) (format.Formatter, error)
	WriterByName    func(string) (writer.Writer, error)
}

type Adapter struct {
	logger          log.Logger
	config          configuration.Provider
	formatterByName func(string) (format.Formatter, error)
	writerByName    func(string) (writer.Writer, error)
}

type EventConfiguration struct {
	InputFormat       string              `config:"inputFormat"`
	OutputTemplate    string              `config:"outputTemplate"`
	OutputWriter      string              `config:"outputWriter"`
	OutputChannel     string              `config:"outputChannel"`
	SingleInputEvent  bool                `config:"singleInputEvent"`
	SingleOutputEvent bool                `config:"singleOutputEvent"`
	ChrootPath        string              `config:"chrootPath"`
	ExtendedFunctions map[string][]string `config:"extendedFunctions"`
	BatchSize         int64               `config:"batchSize"`
	BatchInterval     string              `config:"batchInterval"`
	batchInterval     time.Duration
}

func NewAdapter(cfg *AdapterConfiguration) (*Adapter, error) {
	return &Adapter{
		logger:          cfg.Logger,
		config:          cfg.Config,
		formatterByName: cfg.FormatterByName,
		writerByName:    cfg.WriterByName,
	}, nil
}

func (a *Adapter) AdaptEvent(eventCfg *EventConfiguration) (func([]byte) error, error) {
	writer, err := a.writerByName(eventCfg.OutputWriter)
	if err != nil {
		return nil, err
	}
	formatter, err := a.formatterByName(eventCfg.InputFormat)
	if err != nil {
		return nil, err
	}
	tmpl, err := a.getOutputTemplate(eventCfg)
	if err != nil {
		return nil, err
	}
	channelTmpl, err := a.getChannelTemplate(eventCfg)
	batchInterval := defaultBatchInterval
	if eventCfg.BatchSize > 0 {
		if d, err := time.ParseDuration(eventCfg.BatchInterval); err == nil {
			batchInterval = d
		} else {
			a.logger.Err("error while parsing batch interval (%s) using default batch interval instead (%s)", err.Error(), defaultBatchInterval.String())
		}
		eventCfg.batchInterval = batchInterval
		a.logger.Debug("batch is enable (output channel %s), batch size: %d, batch interval: %s", eventCfg.OutputChannel, eventCfg.BatchSize, eventCfg.batchInterval.String())
	}
	return func(event []byte) error {
		outputs := make(chan (output))
		go a.outputsFromEvent(event, eventCfg, formatter, tmpl, channelTmpl, outputs)
		for o := range outputs {
			if err = writer.Write(o.channel, o.body); err != nil {
				a.logger.Err(err.Error())
			}
		}
		return nil
	}, nil
}

type output struct {
	channel string
	body    []byte
}

func (a *Adapter) outputsFromEvent(event []byte, eventCfg *EventConfiguration, formatter format.Formatter, tmpl *gotemplate.Template, channelTmpl *gotemplate.Template, outputs chan (output)) {
	g := errgroup.Group{}
	var err error
	var elem interface{}
	if eventCfg.SingleInputEvent {
		elem, err = formatter.FormatSingle(event)
		if err != nil {
			a.logger.Err(err.Error())
			close(outputs)
			return
		}
	} else {
		elem, err = formatter.FormatMultiple(event)
		if err != nil {
			a.logger.Err(err.Error())
			close(outputs)
			return
		}
		if len(elem.([]interface{})) == 0 {
			a.logger.Err("input is invalid")
			close(outputs)
			return
		}
	}
	if eventCfg.SingleOutputEvent {
		g.Go(func() error {
			ev, err := a.executeTemplate(tmpl, elem)
			if err != nil {
				return err
			}
			channel, err := a.executeTemplate(channelTmpl, elem)
			if err != nil {
				a.logger.Err(err.Error())
				return err
			}
			outputs <- output{channel: string(channel), body: ev}
			return nil
		})
	} else {
		if !eventCfg.SingleInputEvent {
			elems := elem.([]interface{})
			var elemsPerBatch int64
			for i := 0; i < len(elems); i++ {
				if eventCfg.BatchSize > 0 {
					if eventCfg.BatchSize == elemsPerBatch {
						elemsPerBatch = 0
						time.Sleep(eventCfg.batchInterval)
					}
				}
				i := i
				g.Go(func() error {
					i = i
					ev, err := a.executeTemplate(tmpl, elems[i])
					if err != nil {
						return err
					}
					channel, err := a.executeTemplate(channelTmpl, elems[i])
					if err != nil {
						return err
					}
					outputs <- output{channel: string(channel), body: ev}
					return nil
				})
				if eventCfg.BatchSize > 0 {
					elemsPerBatch += 1
				}
			}
		} else if eventCfg.ChrootPath != "" {
			a.logger.Err("%v: output type invalid: cannot have multiple output type for a single input type if chrootPath is empty", eventCfg)
			close(outputs)
			return
		} else {
			// implement possibility to chroot and iterate through this chrooted key
		}
	}
	if err := g.Wait(); err != nil {
		a.logger.Err(err.Error())
		close(outputs)
		return
	}
}

func (a *Adapter) getOutputTemplate(eventCfg *EventConfiguration) (*gotemplate.Template, error) {
	tmpl := gotemplate.New(filepath.Base(eventCfg.OutputTemplate))
	funcs := template.GetDefaultFuncs()
	if len(eventCfg.ExtendedFunctions) > 0 {
		for pluginFile, functions := range eventCfg.ExtendedFunctions {
			p, err := plugin.Open(pluginFile)
			if err != nil {
				return nil, err
			}
			for _, function := range functions {
				s, err := p.Lookup(function)
				if err != nil {
					return nil, err
				}
				f, ok := s.(func(...interface{}) (interface{}, error))
				if !ok {
					return nil, fmt.Errorf("extended function %s has an invalid signature (required signature is func(...interface{}) (interface{}, error))", function)
				}
				funcs[function] = f
			}
		}
	}
	tmpl, err := tmpl.Funcs(funcs).ParseFiles(eventCfg.OutputTemplate)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (a *Adapter) getChannelTemplate(eventCfg *EventConfiguration) (*gotemplate.Template, error) {
	tmpl := gotemplate.New("channel")
	funcs := template.GetDefaultFuncs()
	if len(eventCfg.ExtendedFunctions) > 0 {
		for pluginFile, functions := range eventCfg.ExtendedFunctions {
			p, err := plugin.Open(pluginFile)
			if err != nil {
				return nil, err
			}
			for _, function := range functions {
				s, err := p.Lookup(function)
				if err != nil {
					return nil, err
				}
				f, ok := s.(func(...interface{}) (interface{}, error))
				if !ok {
					return nil, fmt.Errorf("extended function %s has an invalid signature (required signature is func(...interface{}) (interface{}, error))", function)
				}
				funcs[function] = f
			}
		}
	}
	tmpl, err := tmpl.Funcs(funcs).Parse(eventCfg.OutputChannel)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (a *Adapter) executeTemplate(tmpl *gotemplate.Template, data interface{}) ([]byte, error) {
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return nil, err
	}
	return body.Bytes(), nil
}

func (a *EventConfiguration) ensureConfiguration() error {
	return nil
}
