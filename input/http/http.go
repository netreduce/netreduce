/*
Package http implements a poll/push client/server for http based configuration input sources on-the-fly
update of the definitions.

The poll client (web hook) and the push server provide simple HTTP interfaces that can be implemented by any
third party application.

The HTTP poll interface

The client will regularly poll for definitions on the configured URL. The client assumes that the server
always returns the complete set of definitions that it knows about. Note that Netreduce can be used with
multiple input implementations. The server implementation may or may not support long polling, the client
will wait max until the configured timeout. The server may or may not support the etag or the
If-Modified-Since header, the client will return the previously received values and will consider 304
responses as unchanged input. The client expects the definitions in the response body in application/nred
format (default), or in application/json format (not implemented).

The HTTP push interface

The HTTP input, configured, opens a listener with the provided address to accept incoming definitions from
POST and PUT requests. It returns them for GET requests, or deletes them for DELETE requests. It supports
HEAD and OPTIONS requests. It is possible to segment the definitions with incoming paths. If a set of
definitions was sent previously to a specific path, then GET, HEAD and DELETE will consider the definitions
sent only to that path. In these cases, it is not possible to get back all the definitions with a single
request. The server expects the definitions in the request body in application/nred format (default), or in
application/json format (not implemented).

Note that currently the push interface updates the definitions only in a single instance of Netreduce, so in
a distributed setup, every current instance would need to be udpated, but it is recommended to rather rely
on the poll client in these cases. If it is affordable, it is possible to run an aggregator, that accepts
push requests, and serves as a server for polling by the distributed instances. Currently, there is no
builtin automatic fallback mechanism for the aggregator instances.

Aggregator server

The aggregator server can receive push requests with, or poll for incoming definitions, and can serve
multiple Netreduce instances with the combined set of definitions. This is an experimental feature,
replication is currently not supported, and in case of the outage of the aggregator, especially in the push
scenario, the Netreduce instances may suffer from not getting definition udpates.

Resolving conflicts

The HTTP input implementation sends only the exported definitions to the Netreduce engine. The exported
definitions are identified by their path. The resolution of conflicting exported paths happens by the
Netreduce engine, and the way of handling them is currently undefined.
*/
package http

import (
	"net/http"

	"github.com/netreduce/netreduce"
)

type client struct{}

type Options struct {
	PollTimeout time.Duration
}

func New(o Options) netreduce.Input {
	return &client{}
}

func NewAggregator(o Options) http.Handler {
	return &client{}
}

func (c *client) Get(chan<- netreduce.DefinitionInput) {}

func (c *client) Close() {}

func (c *client) ServeHTTP(http.ResponseWriter, *http.Request) {}
