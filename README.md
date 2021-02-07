### REST-plugin for CoreDNS

This repo contains a plugin for CoreDNS, that enables remote REST service as a source of truth for incoming DNS requests

### How to use?

1. Follow CoreDNS [plugin installation guide](https://coredns.io/2017/07/25/compile-time-enabling-or-disabling-plugins/)
2. Add the plugin config to your `Corefile`
```
.:53 {
    rest https://domain.com/rest/v1/lookup 3600
}
```
With this configuration, once plugin gets lookup request `A example.org`, it will issue HTTP GET requests to the URL `https://domain.co/rest/v1/lookup/A/example.org`

For better performance `cache` plugin will make sense as well.