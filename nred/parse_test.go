package nred

import (
	"bytes"
	"testing"

	"github.com/netreduce/netreduce/data"
)

type parseTestItem struct {
	title      string
	doc        string
	expect     Definition
	expectMany []Definition
	expectNone bool
}

func (test parseTestItem) run(t *testing.T) {
	t.Run(test.title, func(t *testing.T) {
		if test.expectNone {
			testParse(t, test.doc)
		} else if len(test.expectMany) > 0 {
			testParse(t, test.doc, test.expectMany...)
		} else {
			testParse(t, test.doc, test.expect)
		}
	})
}

// destroys the expected slice
func testParse(t *testing.T, doc string, expected ...Definition) {
	d, err := Parse(bytes.NewBufferString(doc))
	if err != nil {
		t.Fatal(err)
	}

	if len(d) != len(expected) {
		t.Error("invalid length of results")
		t.Log("got:     ", len(d))
		t.Log("expected:", len(expected))
		return
	}

	for len(d) > 0 {
		var found bool
		for i := len(expected) - 1; i >= 0; i-- {
			if !Eq(d[0], expected[i]) {
				continue
			}

			found = true
			expected = append(expected[:i], expected[i+1:]...)
			break
		}

		if !found {
			t.Error("definition not found:", d[0].Name())
			t.Log("got:     ", Sprint(d...))
			t.Log("expected:", Sprint(expected...))
		}

		d = d[1:]
	}
}

func TestParse(t *testing.T) {
	for _, test := range []parseTestItem{{
		title:  "hello",
		doc:    `export "/hello" "Hello, world!"`,
		expect: Export("/hello", Definition{}.SetValue(data.String("Hello, world!"))),
	}, {
		title: "pass through",
		doc:   `export "/pass-through" query("https://api.example.org")`,
		expect: Export(
			"/pass-through",
			Definition{}.Query(NewQuery(NewRule("url", "https://api.example.org"))),
		),
	}, {
		title:  "empty",
		doc:    `export "/empty" = define()`,
		expect: Export("/empty", Definition{}),
	}, {
		title: "only local",
		doc: `
			let constants = define(
				const("foo")
				const(42)
				const(3.14)
			)
		`,
		expectNone: true,
	}, {
		title: "reusable local",
		doc: `
			let constants = define(
				const("a", "foo")
				const("b", 42)
				const("c", 3.14)
			)

			export "/constants" constants
		`,
		expect: Export("/constants", Definition{}.Field(
			Const("a", data.String("foo")),
			Const("b", data.Int(42)),
			Const("c", data.Float(3.14)),
		)),
	}, {
		title: "extend",
		doc: `
			let constants = define(
				const("a", "foo")
				const("b", 42)
				const("c", 3.14)
			)

			export "enriched-constants" define(
				constants
				query("https://api.example.org")
				string("foo")
			)
		`,
		expect: Export("enriched-constants", Definition{}.Field(
			Const("a", data.String("foo")),
			Const("b", data.Int(42)),
			Const("c", data.Float(3.14)),
		).Query(
			NewQuery(NewRule("url", "https://api.example.org")),
		).Field(
			String("foo"),
		)),
	}, {
		title: "rules",
		doc: `
			let mapping1 renameField("foo", "bar");
			let mapping2 renameField("bar", "baz");

			export "/foo-bar-baz" define(
				query("https://api.example.org")
				string("foo")
				mapping1
				mapping2
			)
		`,
		expect: Export("/foo-bar-baz", Definition{}.Query(
			NewQuery(NewRule("url", "https://api.example.org")),
		).Field(
			String("foo"),
		).Rule(
			NewRule("renameField", "foo", "bar"),
			NewRule("renameField", "bar", "baz"),
		)),
	}, {
		title: "comments",
		doc: `
			// various mappings are possible, plus custom ones:
			let mapping1 renameField("foo", "bar");
			let mapping2 renameField("bar", "baz");

			// the backend returns {"foo": "blah"}
			// it's renamed to {"bar": "blah"}
			// then to {"baz": "blah"}
			export "/foo-bar-baz" define(
				query("https://api.example.org")
				string("foo")
				mapping1
				mapping2

				// we don't need this here, just demoing comments:
				/* int("qux") */
			)
		`,
		expect: Export("/foo-bar-baz", Definition{}.Query(
			NewQuery(NewRule("url", "https://api.example.org")),
		).Field(
			String("foo"),
		).Rule(
			NewRule("renameField", "foo", "bar"),
			NewRule("renameField", "bar", "baz"),
		)),
	}, {
		title: "contains",
		doc: `
			export "/authenticated-user" define(
				query("https://auth.example.org/info")
				query(authConnector.extended)

				string("name")
				int("level")
				float("iris-radius-when-seen-this")

				contains("roles", define(
					query("https://auth.example.org/roles", path(link("id")))
					string("name")
					selectField("name")
				))
			)
		`,
		expect: Export("/authenticated-user", Definition{}.Query(
			NewQuery(NewRule("url", "https://auth.example.org/info")),
			NewQuery(NewRule("authConnector.extended")),
		).Field(
			String("name"),
			Int("level"),
			Float("iris-radius-when-seen-this"),
			Contains("roles", Definition{}.Query(
				NewQuery(
					NewRule("url", "https://auth.example.org/roles"),
					NewRule("path", NewRule("link", "id")),
				),
			).Field(
				String("name"),
			).Rule(
				NewRule("selectField", "name"),
			)),
		)),
	}} {
		test.run(t)
	}
}

func TestParseRef(t *testing.T) {
	for _, test := range []parseTestItem{{
		title:  "value as define",
		doc:    `export "/hello" define("Hello, world!")`,
		expect: Export("/hello", Definition{}.SetValue(data.String("Hello, world!"))),
	}, {
		title: "const as definition",
		doc: `
			let foo = const("foo", 42)
			export "/foo" foo
		`,
		expect: Export("/foo", Definition{}.Field(Const("foo", data.Int(42)))),
	}, {
		title:  "empty define",
		doc:    `export "empty" define`,
		expect: Export("empty", Definition{}),
	}, {
		title:  "rule without args",
		doc:    `export "foo" bar`,
		expect: Export("foo", Definition{}.Rule(NewRule("bar"))),
	}, {
		title:  "curry define",
		doc:    `export "foo" define(const("a", 1))(const("b", 2))`,
		expect: Export("foo", Definition{}.Field(Const("a", data.Int(1)), Const("b", data.Int(2)))),
	}, {
		title: "primitive as initial definition",
		doc:   `export "one-foo" 1(const("foo", 42))`,
		expect: Export("one-foo", Definition{}.SetValue(
			data.Int(1),
		).Field(
			Const("foo", data.Int(42)),
		)),
	}, {
		title: "primitive as initial definition, explained",
		doc:   `export "one-foo" define(1)(define(const("foo", 42)))`,
		expect: Export("one-foo", Definition{}.SetValue(
			data.Int(1),
		).Field(
			Const("foo", data.Int(42))),
		),
	}, {
		title: "primitive as initial definition, simplified",
		doc:   `export "one-foo" define(1, const("foo", 42))`,
		expect: Export("one-foo", Definition{}.SetValue(
			data.Int(1),
		).Field(
			Const("foo", data.Int(42)),
		)),
	}} {
		test.run(t)
	}
}
