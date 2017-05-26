package enforcer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
)

type work struct {
	r  *http.Request
	w  http.ResponseWriter
	ch chan bool

	wg sync.WaitGroup

	callNext   bool
	httpStatus int
	httpErr    error
}

func newWork(w http.ResponseWriter, r *http.Request) *work {
	return &work{
		ch: make(chan bool),
		r:  r,
		w:  w,
	}
}

func (w *work) wait() {
	<-w.ch
}

func (w *work) complete(next bool, status int, err error) {
	w.callNext = next
	w.httpStatus = status
	w.httpErr = err
	close(w.ch)
}

type client struct {
	url string
}

func newClient(url string) *client {
	return &client{
		url: url,
	}
}

func (c *client) serve(ch <-chan *work) {
	for w := range ch {
		c.handle(w)
	}
}

func (c *client) handle(w *work) {
	hr := RequestFor(w.r)
	r, err := buildRequest(c.url, hr)
	if err != nil {
		w.httpErr = err
		return
	}
	r = r.WithContext(w.r.Context())

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		w.complete(false, 0, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// what to do when we are down?
		w.complete(true, 0, nil)
		return
	}

	hresp := &Response{}
	if err = json.NewDecoder(resp.Body).Decode(&hresp); err != nil {
		w.complete(false, 0, err)
		return
	}

	w.complete(hresp.Resolve(w.w, w.r))
}

func buildRequest(url string, r *Request) (*http.Request, error) {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(r)
	if err != nil {
		return nil, err
	}

	return http.NewRequest("GET", url, buf)
}
