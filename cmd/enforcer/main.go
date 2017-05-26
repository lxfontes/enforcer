package main

import (
	"github.com/mholt/caddy/caddy/caddymain"
	"github.com/mholt/caddy/caddyhttp/httpserver"

	_ "github.com/captncraig/caddy-realip"
	"github.com/lxfontes/enforcer"
)

func main() {
	httpserver.RegisterDevDirective(enforcer.CaddyDirective, "proxy")
	caddymain.Run()
}
