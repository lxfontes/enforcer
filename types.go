package enforcer

import "net/http"

type Request struct {
	URL     string              `json:"url"`
	Host    string              `json:"host"`
	Headers map[string][]string `json:"headers"`
}

type Response struct {
	AppendHeaders map[string]string `json:"append_headers"`

	Content struct {
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
	}

	RemoveHeaders []string `json:"remove_headers"`

	Status int `json:"status"`
}

func RequestFor(r *http.Request) *Request {
	req := &Request{
		Host:    r.Host,
		URL:     r.URL.String(),
		Headers: make(map[string][]string),
	}

	for k, v := range r.Header {
		req.Headers[k] = v
	}

	return req
}

func (resp *Response) Resolve(w http.ResponseWriter, r *http.Request) (bool, int, error) {
	if len(resp.RemoveHeaders) > 0 {
		for _, k := range resp.RemoveHeaders {
			r.Header.Del(k)
		}
	}

	if len(resp.AppendHeaders) > 0 {
		for k, v := range resp.AppendHeaders {
			r.Header.Set(k, v)
		}
	}

	next := true
	if resp.Status > 0 {
		w.WriteHeader(resp.Status)
		next = false
	}

	if resp.Content.Body != "" {
		for k, v := range resp.Content.Headers {
			w.Header().Set(k, v)
		}

		if _, err := w.Write([]byte(resp.Content.Body)); err != nil {
			return false, 0, err
		}
		next = false
	}

	return next, 0, nil
}
