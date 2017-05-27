package enforcer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

const backOffInitial = 500 * time.Millisecond
const backOffMultiplier = 1.5
const backOffMax = 5 * time.Second

type work struct {
	r  *http.Request
	w  http.ResponseWriter
	ch chan bool

	resp *Response
	err  error
}

func newWork(w http.ResponseWriter, r *http.Request) *work {
	return &work{
		ch:   make(chan bool, 1),
		r:    r,
		w:    w,
		resp: &Response{},
	}
}

func (w *work) serveHTTP(rw http.ResponseWriter, r *http.Request, nextHandler httpserver.Handler) (int, error) {
	_, ok := <-w.ch
	if !ok {
		return http.StatusInternalServerError, ErrWorkState
	}

	close(w.ch)
	if w.err != nil {
		return http.StatusInternalServerError, w.err
	}

	return w.resp.Resolve(rw, r, nextHandler)
}

func (w *work) notify() {
	w.ch <- true
}

type client struct {
	url    string
	wcount int
}

func newClient(url string, workers int) *client {
	return &client{
		url:    url,
		wcount: workers,
	}
}

func (c *client) serve(ch <-chan *work) {
	var wg sync.WaitGroup

	for i := 0; i < c.wcount; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			backOffCycle := 0

			for w := range ch {
				if err := c.handle(w); err != nil {
					backOffCycle = backOff(backOffCycle)
				} else {
					backOffCycle = 0
				}
			}
		}()
	}

	wg.Wait()
}

func backOff(cycle int) int {
	timeout := time.Duration((float32(cycle)*backOffMultiplier))*backOffInitial + backOffInitial
	if timeout > backOffMax {
		timeout = backOffInitial
		cycle = 0
	}

	<-time.After(timeout)

	return cycle + 1
}

func (c *client) handle(w *work) error {
	defer w.notify()

	hr := RequestFor(w.r)
	r, err := buildRequest(c.url, hr)
	if err != nil {
		w.err = err
		return nil
	}
	r = r.WithContext(w.r.Context())

	// start backoff in case:
	// - rule engine is not available (net timeout / conn refused)
	// - rule engine returns status != 200
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		w.err = err
		return w.err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.err = ErrInvalidServerStatus
		return w.err
	}

	if err = json.NewDecoder(resp.Body).Decode(&w.resp); err != nil {
		w.err = err
	}

	return nil
}

func buildRequest(url string, r *Request) (*http.Request, error) {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(r)
	if err != nil {
		return nil, err
	}

	return http.NewRequest("GET", url, buf)
}
