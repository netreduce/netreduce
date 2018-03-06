/*
Package http implements a pull/push client/server for http based configuration input sources on-the-fly update
of the definitions.

The HTTP pull interface (web hook)

The client will poll for definitions on the configured URL. The client assumes that the server always returns
the complete set of definitions that it knows about. Note that there netreduce can be used with multiple input
implementations. The server implementation may or may not support long polling, the client will wait max until
the configured timeout. The server may or may not support the etag or the If-Modified-Since header, the client
will return the previously received values and will consider 304 responses as unchanged input. The client
expects the definitions in the response body in application/nred format (default), or in application/json
format (not implemented).

The HTTP push interface

The HTTP input, configured, opens a listener with the provided address to accept incoming definitions from POST
and PUT requests. It returns them for GET requests, or deletes them for DELETE requests. It supports HEAD and
OPTIONS requests. It is possible to segment the definitions with incoming paths. If a set of definitions was
sent previously to a specific path, then GET, HEAD and DELETE will consider the definitions sent to that path.
The server expects the definitions in the request body in application/nred format (default), or in
application/json format (not implemented).

Note that currently the push interface updates the definitions only in a single instance, so in a distributed
setup, every current instance would need to be udpated, but it is recommended to rather rely on the pull client
in these cases.

Resolving Conflicts

The HTTP input implementation sends only the exported definitions to the netreduce engine. The exported
definitions are identified by their path. The resolution of conflicting exported paths happens by the netreduce
engine, and the way of handling them is currently undefined.
*/
package http

type client struct {}

type Options struct {
	PollTimeout time.Duration
}

func New(o Options) netreduce.Input {
	return client
}

func (c *client) Get(chan<- netreduce.DefinitionInput) {}

func (c *client) Close() {}
