# F-PKI Map Server Plugin

## Map Server Plugin

Fetch F-PKI map server proof bytes via DNS

## Description

This plugin enables the DNS resolver to return F-PKI map server proof bytes via DNS responses in a TXT resource record.

The configuration looks as follows:

~~~ corefile
mapserver-domain {
  mapserver path-to-mapserver-id path-to-mapserver-public-key mapserver-grpc-address max-receive-message-size
}
~~~

## Compilation

This package will always be compiled as part of CoreDNS and not in a standalone way. It will require you to use `go get` or as a dependency on [plugin.cfg](https://github.com/coredns/coredns/blob/master/plugin.cfg).

The [manual](https://coredns.io/manual/toc/#what-is-coredns) will have more information about how to configure and extend the server with external plugins.

A simple way to consume this plugin, is by adding the following on [plugin.cfg](https://github.com/coredns/coredns/blob/master/plugin.cfg), and recompile it as [detailed on coredns.io](https://coredns.io/2017/07/25/compile-time-enabling-or-disabling-plugins/#build-with-compile-time-configuration-file).

~~~
example:github.com/coredns/example
~~~

After this you can compile coredns by:

```shell script
go generate
go build
```

Or you can instead use make:

```shell script
make
```

## Examples

In this configuration, we forward all queries to 9.9.9.9 and print "example" whenever we receive
a query.

~~~ corefile
mapserver1.com:12345 {
  mapserver /path/to/mapid1 /path/to/mappk1.pem grpc://localhost:8094 1073741824
}
~~~
