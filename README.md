# http-event-adapter

Generic way to send business events inside sink
It simply listen from HTTP requests and using gotemplate, transform request, and send the output data inside a sink.
All is configurable, contribution opened :)

## Supported format

- csv
- yaml
- json

## Writers

- fmt
- nats (to do)

## Examples

```
curl --data-binary "@examples/example.json" -X POST http://localhost:8081/json
curl --data-binary "@examples/example.csv" -X POST http://localhost:8081/csv
curl --data-binary "@examples/example.yaml" -X POST http://localhost:8081/yaml
```

those requests use examples/person.tmpl as template (see configuration below)

## Configuration

```go
type EventConfiguration struct {
	InputFormat       string              `config:"inputFormat"` // (json/yaml/csv)
	OutputTemplate    string              `config:"outputTemplate"` // (path of the template file)
	OutputWriter      string              `config:"outputWriter"` (fmt/nats)
	OutputChannel     string              `config:"outputChannel"` (the path of the output channel / can be templatize)
	SingleOutputEvent bool                `config:"singleOutputEvent"` (set true if one http request == one output event, default to false)
	ChrootPath        string              `config:"chrootPath"` (not implemented)
	ExtendedFunctions map[string][]string `config:"extendedFunctions"` (specify informations for extended functions. key is link for the so file, values are all exporter functions you want to use in templates)
}
```

## TODO

- one separator per csv route
