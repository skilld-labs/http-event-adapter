# Go templates extension

http-event-adapter use go plugins (https://golang.org/pkg/plugin/) to be able to extend go templates, by providing custom functions that can be call inside a go template.

## Usage

Only thing to do is to create directory and write a `main.go` inside, extended function has to be exported.
Then you can build .so file with `go build -buildmode=plugin`. It will create a .so file with the same name than the directory.

To use this .so file, you have to add `extendedFunctions` part inside `config.yaml`.

eg:

```
  /csv:
    inputFormat: csv
    outputTemplate: examples/person.tmpl
    outputWriter: fmt
    outputChannel: /person
    extendedFunctions:
      examples/extensions/username/username.so:
        - Username
```
