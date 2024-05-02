package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/skilld-labs/http-event-adapter/configuration"
	"github.com/skilld-labs/http-event-adapter/log"

	"github.com/skilld-labs/http-event-adapter/adapter"
	"github.com/skilld-labs/http-event-adapter/format"
	"github.com/skilld-labs/http-event-adapter/router"
	"github.com/skilld-labs/http-event-adapter/writer"
)

func main() {
	config := flag.String("config", "config.yaml", "The path for the config file")
	flag.Parse()

	l := log.NewJsonLogger(&log.LoggerConfiguration{})
	cfg, err := configuration.NewKoanfProvider(configuration.ProviderConfig{
		Logger: l,
		Source: *config,
	})
	if err != nil {
		l.Fatal(err.Error())
	}
	l.SetVerbosity(cfg.GetString("log.verbosity"))

	r := router.NewRouter(&router.RouterConfiguration{Logger: l})

	formatterCfg := &format.FormatterConfiguration{Logger: l, Config: cfg}
	writerCfg := &writer.WriterConfiguration{Logger: l, Config: cfg}
	a, err := adapter.NewAdapter(&adapter.AdapterConfiguration{
		Logger: l,
		Config: cfg,
		FormatterByName: func(name string) (format.Formatter, error) {
			return format.GetFormatter(formatterCfg, name)
		},
		WriterByName: func(name string) (writer.Writer, error) {
			return writer.GetWriter(writerCfg, name)
		},
	})
	if err != nil {
		l.Fatal(err.Error())
	}
	events := make(map[string]*adapter.EventConfiguration)
	cfg.Load("events", &events)
	if len(events) == 0 {
		l.Fatal("no event configuration ... exiting")
	}
	for path, event := range events {
		c, err := a.AdaptEvent(event)
		if err != nil {
			l.Fatal(err.Error())
		}
		r.AddRoute(path, c)
	}

	http.Handle("/", debugMiddleware(l, cfg.GetBool("debug"), cfg.GetString("debugDirectory"), r))

	port := cfg.GetString("port")
	l.Info("server listening on %s", port)
	l.Fatal(http.ListenAndServe(":"+port, nil).Error())
}

func debugMiddleware(logger log.Logger, debug bool, debugDirectory string, handler http.Handler) http.Handler {
	if !debug {
		logger.Debug("debug mode is disabled")
		return handler
	}
	logger.Debug("debug mode is activated")
	// Create a directory if it doesn't exist
	if _, err := os.Stat(debugDirectory); os.IsNotExist(err) {
		logger.Debug("creating directory %s for debug purpose", debugDirectory)
		os.MkdirAll(debugDirectory, 0755) // Adjust the permissions as needed
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Read the body data
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}

		// Restore the body so it can be read by the original handler
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		// Generate a filename based on the current time to avoid overwrites
		timestamp := time.Now().Format("20060102150405.999999")
		fileName := filepath.Join(debugDirectory, timestamp+"_request.txt")

		// Write the body to a file
		err = ioutil.WriteFile(fileName, body, 0644)
		if err != nil {
			http.Error(w, "Error writing request to debug file", http.StatusInternalServerError)
			return
		}
		logger.Debug("file %s has been created for debug purpose", fileName)

		// Call the original handler
		handler.ServeHTTP(w, req)
	})
}
