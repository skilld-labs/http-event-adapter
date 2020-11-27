package configuration

import (
	"github.com/skilld-labs/http-event-adapter/log"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
)

const delimiter = "."

type provider struct {
	logger log.Logger
	*koanf.Koanf
}

func NewKoanfProvider(c ProviderConfig) (Provider, error) {
	k := koanf.New(delimiter)
	switch source := c.Source.(type) {
	case string:
		if err := k.Load(file.Provider(source), yaml.Parser()); err != nil {
			return nil, err
		}
	case *provider:
		k.Merge(source.Koanf)
	}
	return &provider{logger: c.Logger, Koanf: k}, nil
}

func (p *provider) GetBool(key string) bool {
	return p.Bool(key)
}

func (p *provider) GetFloat(key string) float64 {
	v := p.Float64(key)
	if v == 0 {
		p.logger.Debug("%s not found", key)
	}
	return v
}

func (p *provider) GetFloats(key string) []float64 {
	v := p.Float64s(key)
	if len(v) == 0 {
		p.logger.Debug("%s not found", key)
	}
	return v
}

func (p *provider) GetString(key string) string {
	v := p.String(key)
	if v == "" {
		p.logger.Debug("%s not found", key)
	}
	return v
}

func (p *provider) GetStrings(key string) []string {
	v := p.Strings(key)
	if len(v) == 0 {
		p.logger.Debug("%s not found", key)
	}
	return v
}

func (p *provider) GetMapStringString(key string) map[string]string {
	v := p.StringMap(key)
	if len(v) == 0 {
		p.logger.Debug("%s not found", key)
	}
	return v
}

func (p *provider) GetMapStringStrings(key string) map[string][]string {
	v := p.StringsMap(key)
	if len(v) == 0 {
		p.logger.Debug("%s not found", key)
	}
	return v
}

func (p *provider) GetMapStringBool(key string) map[string]bool {
	v := p.BoolMap(key)
	if len(v) == 0 {
		p.logger.Debug("%s not found", key)
	}
	return v
}

func (p *provider) Load(key string, ptr interface{}) {
	if err := p.Koanf.UnmarshalWithConf(key, ptr, koanf.UnmarshalConf{Tag: Tag}); err != nil {
		p.logger.Err("error while loading %s : %s", key, err.Error())
	}
}

func (p *provider) Set(key string, value interface{}) {
	p.Koanf.Load(confmap.Provider(map[string]interface{}{key: value}, delimiter), nil)
}
