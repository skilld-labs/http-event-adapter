package router

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/skilld-labs/http-event-adapter/log"
)

type RouterConfiguration struct {
	Logger log.Logger
}

type Router struct {
	logger log.Logger
	routes map[string]func([]byte)
}

func NewRouter(cfg *RouterConfiguration) *Router {
	return &Router{logger: cfg.Logger, routes: make(map[string]func([]byte))}
}

func (r *Router) AddRoute(path string, callback func([]byte)) error {
	if _, exists := r.routes[path]; exists {
		return fmt.Errorf("a route have been already registered on route %s", path)
	}
	r.routes[path] = callback
	r.logger.Debug("added new route on path %s", path)
	return nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.logger.Debug("received a new request on %s", req.URL.Path)
	if req.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		r.logger.Err("method %s is not allowed", req.Method)
		return
	}
	if callback, exists := r.routes[req.URL.Path]; exists {
		fmt.Fprint(w, "ok")
		body := new(bytes.Buffer)
		body.ReadFrom(req.Body)
		callback(body.Bytes())
	} else {
		r.logger.Err("no route on path %s", req.URL.Path)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}
