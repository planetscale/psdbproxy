# psdbproxy

A local mysql proxy for interacting with a PlanetScale database over HTTP.

This proxy terminates a mysql connection on a local socket, and multiplexes all connections over an HTTP/2 stream to PlanetScale.

Currently is a WIP, but intended to be utilized as a library as well as a standalone tool.
