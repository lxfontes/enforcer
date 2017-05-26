See [Caddyfile](cmd/enforcer/Caddyfile) for example.

```
enforcer {
  http://enforcer1/engine
  https://enforcer2/engine
}
```

enforcer talks to an external rules engine and manipulate/answer the current request.

sample request

```
GET /ratelimit HTTP/1.1
Host: localhost:7070
User-Agent: Go-http-client/1.1
Content-Length: 101
Accept-Encoding: gzip

{"url":"/headers","host":"localhost:2015","headers":{"Accept":["*/*"],"User-Agent":["curl/7.51.0"]}}
```

sample response, manipulating headers:
```
HTTP/1.0 200 OK
Content-type: application/json

{
  "append_headers": {
    "X-That": "bla"
  },
  "remove_headers": [
    "User-Agent"
  ]
}
```

sample response, answering request:
```
HTTP/1.0 200 OK
Content-type: application/json

{
  "content": {
    "headers": {
      "X-This": "bla"
    },
    "body": "go away"
  },
  "status": 403
}
```
