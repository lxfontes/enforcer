package enforcer

import (
	"net/url"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// CaddyDirective ...
const CaddyDirective = "enforcer"

func init() {
	caddy.RegisterPlugin(CaddyDirective, caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	var e *enforcer
	for c.Next() {
		if e != nil {
			return c.Err("cannot specify enforcer more than once")
		}

		var serverAddrs []string

		hadBlock := false
		for c.NextBlock() {
			hadBlock = true

			addr := c.Val()
			if _, err := url.Parse(addr); err != nil {
				return c.ArgErr()
			}
			serverAddrs = append(serverAddrs, addr)

			if c.NextArg() {
				return c.ArgErr()
			}
		}

		if !hadBlock {
			return c.Err("expecting block")
		}

		var err error
		e, err = newEnforcer(serverAddrs)
		if err != nil {
			return err
		}
	}

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		e.next = next
		return e
	})

	c.OnShutdown(e.stop)

	return nil
}
