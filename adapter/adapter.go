package adapter

import (
	"bytes"
	"fmt"

	"path/filepath"
	"plugin"
	"reflect"
	gotemplate "text/template"

	"github.com/skilld-labs/http-event-adapter/log"

	"github.com/skilld-labs/http-event-adapter/configuration"

	"github.com/skilld-labs/http-event-adapter/format"
	"github.com/skilld-labs/http-event-adapter/template"
	"github.com/skilld-labs/http-event-adapter/writer"
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
	SingleOutputEvent bool                `config:"singleOutputEvent"`
	ChrootPath        string              `config:"chrootPath"`
	ExtendedFunctions map[string][]string `config:"extendedFunctions"`
}

func NewAdapter(cfg *AdapterConfiguration) (*Adapter, error) {
	return &Adapter{
		logger:          cfg.Logger,
		config:          cfg.Config,
		formatterByName: cfg.FormatterByName,
		writerByName:    cfg.WriterByName,
	}, nil
}

func (a *Adapter) AdaptEvent(eventCfg *EventConfiguration) (func([]byte), error) {
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
	return func(event []byte) {
		oo, err := a.outputsFromEvent(event, eventCfg, formatter, tmpl, channelTmpl)
		if err != nil {
			a.logger.Err(err.Error())
		}
		for _, o := range oo {
			if err = writer.Write(o.channel, o.body); err != nil {
				a.logger.Err(err.Error())
			}
		}
	}, nil
}

type output struct {
	channel string
	body    []byte
}

func (a *Adapter) outputsFromEvent(event []byte, eventCfg *EventConfiguration, formatter format.Formatter, tmpl *gotemplate.Template, channelTmpl *gotemplate.Template) ([]*output, error) {
	var ee []*output
	elem, err := formatter.Format(event)
	if err != nil {
		return nil, err
	}
	if eventCfg.SingleOutputEvent {
		ev, err := a.executeTemplate(tmpl, elem)
		if err != nil {
			return nil, err
		}
		channel, err := a.executeTemplate(channelTmpl, elem)
		if err != nil {
			return nil, err
		}
		ee = append(ee, &output{channel: string(channel), body: ev})
	} else {
		re := reflect.ValueOf(elem)
		if re.Kind() == reflect.Slice {
			for i := 0; i < re.Len(); i++ {
				data := re.Index(i).Interface()
				ev, err := a.executeTemplate(tmpl, data)
				if err != nil {
					return nil, err
				}
				channel, err := a.executeTemplate(channelTmpl, data)
				if err != nil {
					return nil, err
				}
				ee = append(ee, &output{channel: string(channel), body: ev})
			}
		} else if eventCfg.ChrootPath != "" {
			return nil, fmt.Errorf("%v: output type invalid: cannot have multiple output type for a single input type if chrootPath is empty", eventCfg)
		} else {
			// implement possibility to chroot and iterate through this chrooted key
		}
	}
	return ee, nil
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
