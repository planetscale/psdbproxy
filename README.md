# psdbproxy

A local MySQL proxy for interacting with a PlanetScale database over HTTP.

This proxy terminates a MySQL connection on a local socket, and multiplexes all connections over an HTTP/2 stream to PlanetScale.

Currently is a WIP, but intended to be utilized as a library as well as a standalone tool.

## CLI Usage

```
Usage of psdbproxy:
  -C, --compress string      compress traffic with given algorithm (identity, gzip, s2) (default "s2")
  -h, --host string          upstream PlanetScale hostname (default "aws.connect.psdb.cloud")
  -l, --listen string        mysql address to listen (default "127.0.0.1:3306")
  -p, --password string      PlanetScale password
  -u, --username string      PlanetScale username
  -F, --wire-format string   transport wire format (protobuf, json) (default "protobuf")
```
