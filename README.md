# Netreduce

Netreduce is an API aggregator for HTTP services. Its primary goal is to provide an adapter layer between
multiple different backend services and their clients. Netreduce provides an interface that can be optimized
for the requirements of the clients, while allowing a clean and normalized interface on the backend services
that are the owners of the actual resources.

The most abstract, general use case of netreduce is the transformation of service topology. It makes it
possible to keep fulfilling the requirements of the service clients, while allowing the restructuring of the
original sources, or the other way around. Some practical, more concrete example is the BFF, Backend For
Frontend.

### Features

**WIP: netreduce is a work-in-progress project, the below features are meant as currently planned and can be
in different state of availability, can be changed, and finally also can be dropped, until the first beta
version of netreduce is released.**

- many-to-many relation between backend services and frontend endpoints
- free composability of the frontend resource structures and their fields 
- automatic parallelization/optimization of backend requests
- runtime definition of frontend endpoints without downtime
- extensibility with custom backend connectors
- custom mapping functions for the frontend resources
- metrics and tracing

### Examples:

```
// hello:
export "/hello" "Hello, world!"

// pass through:
export "/original" query("https://api.example.org")

// empty, returns {} (equals, and semicolons are optional):
export "/empty" = define()

// local, reusable definition (commas are optional if there are newlines between the args):
let constants = define(
	const("foo")
	const("42")
	const(3.14)
)

export "constants" constants

// only the defined fields are returned by default plus the ID if exists:
export "enriched-constants" extend(
	constants
	query("https://api.example.org")
	string("foo")
)

// various mappings are possible, plus custom ones:
let mapping1 map(renameField("foo", "bar"));
let mapping2 map(renameField("bar", "baz"));

// the backend returns {"foo": "blah"}
// the query maps it to {"bar": "blah"}
// the definition maps it to {"baz": "blah"}
export "/foo-bar-baz" define(
	query("https://api.example.org", mapping1)
	string("bar")
	mapping2

	// we don't need this here, just demoing comments:
	/* int("qux") */
)

// all kinds of relations are possible. It must be a tree, but the backend queries are parallelized and
// deduplicated.
// backend URLs: custom connectors are allowed, but by default, string urls are just automatically wrapped with a
// default http connector.
export "/authenticated-user" define(
	query("https://auth.example.org/info")
	query(authConnector.extended)

	string("name")
	int("level")
	float("iris-radius-when-seen-this :)")

	containsMany("roles", by("id"), define(
		query("https://auth.example.org/roles")
		string("role")
		selectField("role")
	))
)
```
