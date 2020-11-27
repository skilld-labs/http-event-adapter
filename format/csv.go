package format

import (
	"bytes"
	"encoding/csv"
	"io"
)

const (
	defaultSeparator = ','
)

type csvFormatter struct {
	separator rune
}

func NewCsvFormatter(cfg *FormatterConfiguration) (Formatter, error) {
	separator := defaultSeparator
	if sep := cfg.Config.GetString("csv.separator"); sep != "" {
		separator = rune(sep[0])
	}
	return &csvFormatter{separator: separator}, nil
}

func (c *csvFormatter) Format(data []byte) (interface{}, error) {
	r := bytes.NewReader(data)
	return c.csvToMaps(r)
}

func (c *csvFormatter) csvToMaps(reader io.Reader) ([]map[string]string, error) {
	r := csv.NewReader(reader)
	r.Comma = c.separator
	rows := []map[string]string{}
	var header []string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if header == nil {
			header = record
		} else {
			dict := map[string]string{}
			for i := range header {
				dict[header[i]] = record[i]
			}
			rows = append(rows, dict)
		}
	}
	return rows, nil
}
