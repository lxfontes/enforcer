package enforcer

import (
	"net/http"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

type enforcer struct {
	next    httpserver.Handler
	queue   chan *work
	clients []*client
}

func newEnforcer(serverAddrs []string) (*enforcer, error) {
	e := &enforcer{
		queue: make(chan *work),
	}

	for _, addr := range serverAddrs {
		c := newClient(addr)

		e.clients = append(e.clients, c)
	}

	for _, c := range e.clients {
		go c.serve(e.queue)
	}

	return e, nil
}

func (e *enforcer) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	wrk := newWork(w, r)

	e.queue <- wrk
	wrk.wait()

	if wrk.callNext {
		return e.next.ServeHTTP(w, r)
	}

	return wrk.httpStatus, wrk.httpErr
}

func (e *enforcer) stop() error {
	close(e.queue)
	return nil
}
