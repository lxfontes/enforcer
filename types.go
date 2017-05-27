package enforcer

import (
	"errors"
	"net/http"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

var (
	ErrWorkState           = errors.New("worker state is not valid")
	ErrRuleTimeout         = errors.New("failed to communicate with rule engine")
	ErrInvalidServerStatus = errors.New("invalid http status from server")
)

type Request struct {
	// Method is the request method
	Method string `json:"method"`
	// URI is the unmodified URI from Request Line (RFC 2616, Section 5.1)
	URI string `json:"uri"`
	// Host as per RFC 2616 might be the 'Host' header or the server endpoint address
	Host string `json:"host"`
	// Headers contain all client headers 'as-is' from request
	Headers map[string][]string `json:"headers"`
}

type Response struct {
	// SetHeaders sets http request headers.
	// If a header with the same name already exists, it will be overwritten.
	SetHeaders map[string]string `json:"set_headers"`
	// RemoveHeaders removes http request headers.
	RemoveHeaders []string `json:"remove_headers"`

	// Serve contains fields that when present will stop the middleware chain
	Serve struct {
		// Status sets the http response status, stopping middleware chain
		Status int `json:"status"`
		// Headers add headers to http response
		Headers map[string]string `json:"headers"`
		// Body sets the http response body
		Body string `json:"body"`
	} `json:"serve"`
}

func RequestFor(r *http.Request) *Request {
	req := &Request{
		Host:    r.Host,
		URI:     r.RequestURI,
		Headers: make(map[string][]string),
	}

	for k, v := range r.Header {
		req.Headers[k] = v
	}

	return req
}

func (resp *Response) Resolve(w http.ResponseWriter, r *http.Request, nextHandler httpserver.Handler) (int, error) {
	if len(resp.RemoveHeaders) > 0 {
		for _, k := range resp.RemoveHeaders {
			r.Header.Del(k)
		}
	}

	if len(resp.SetHeaders) > 0 {
		for k, v := range resp.SetHeaders {
			r.Header.Set(k, v)
		}
	}

	next := true
	if resp.Serve.Status > 0 {
		w.WriteHeader(resp.Serve.Status)
		next = false
	}

	if resp.Serve.Body != "" {
		for k, v := range resp.Serve.Headers {
			w.Header().Set(k, v)
		}

		if _, err := w.Write([]byte(resp.Serve.Body)); err != nil {
			return 0, err
		}
		next = false
	}

	if !next {
		return 0, nil
	}

	return nextHandler.ServeHTTP(w, r)
}
