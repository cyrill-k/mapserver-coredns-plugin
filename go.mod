module github.com/cyrill-k/mapserver-coredns-plugin

go 1.13

replace github.com/cyrill-k/fpki => /home/cyrill/go/src/github.com/cyrill-k/fpki

require (
	github.com/caddyserver/caddy v1.0.4
	github.com/coredns/coredns v1.6.7
	github.com/cyrill-k/fpki v0.0.0-00010101000000-000000000000
	github.com/miekg/dns v1.1.27
	github.com/prometheus/client_golang v1.4.0
)
