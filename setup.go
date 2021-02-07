package rest

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"regexp"
	"strconv"
)

var reHttp = regexp.MustCompile("(?i)^(http|https)://")

// init registers this plugin.
func init() { plugin.Register("rest", setup) }

func setupPlugin(c *caddy.Controller) (Client, error) {
	var url string
	var ttl uint32
	if c.NextArg() {
		// it must be an url, either HTTP or HTTP
		args := c.RemainingArgs()
		// url must be present anyway
		if len(args) < 1 || !reHttp.MatchString(args[0]) {
			return Client{}, plugin.Error("rest", c.ArgErr())
		}

		url = args[0]

		// ttl is optional
		if len(args) > 1 {
			v, err := strconv.Atoi(args[1])
			if err != nil {
				return Client{}, plugin.Error("rest", err)
			}
			ttl = uint32(v)
		} else {
			ttl = 300
		}
	} else {
		return Client{}, plugin.Error("rest", c.ArgErr())
	}

	return BuildClient(url, ttl)
}

func setup(c *caddy.Controller) error {
	client, err := setupPlugin(c)
	if err != nil {
		return plugin.Error("rest", err)
	}

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return client
	})

	// All OK, return a nil error.
	return nil
}
