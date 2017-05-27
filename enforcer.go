package enforcer

import (
	"net/http"
	"time"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

type enforcer struct {
	next    httpserver.Handler
	queue   chan *work
	clients []*client

	maxWait time.Duration
}

func newEnforcer(serverAddrs []string) (*enforcer, error) {
	e := &enforcer{
		queue:   make(chan *work),
		maxWait: 1 * time.Second,
	}

	for _, addr := range serverAddrs {
		c := newClient(addr, 1)

		e.clients = append(e.clients, c)
	}

	for _, c := range e.clients {
		go c.serve(e.queue)
	}

	return e, nil
}

func (e *enforcer) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	wrk := newWork(w, r)

	select {
	case <-time.After(e.maxWait):
		// discard request
		return http.StatusServiceUnavailable, ErrRuleTimeout
	case e.queue <- wrk:
	}

	return wrk.serveHTTP(w, r, e.next)
}

func (e *enforcer) stop() error {
	close(e.queue)
	return nil
}
