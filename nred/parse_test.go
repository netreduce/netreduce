package nred

import (
	"bytes"
	"testing"
)

// destroys expected slice
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
			expected = append(expected[:i], expected[i + 1:]...)
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
	for _, test := range []struct{
		title string
		doc string
		expected Definition
		expectedMany []Definition
		noneExpected bool
	}{{
		title: "hello",
		doc: `export "/hello" "Hello, world!"`,
		expected: Export("/hello", Define("Hello, world!")),
	}, {
		title: "pass through",
		doc: `export "/pass-through" query("https://api.example.org")`,
		expected: Export("/pass-through", Define(Query(Rule("url", "https://api.example.org")))),
	}, {
		title: "empty",
		doc: `export "/empty" = define()`,
		expected: Export("/empty", Define()),
	}, {
		title: "only local",
		doc: `
			let constants = define(
				const("foo")
				const(42)
				const(3.14)
			)
		`,
		noneExpected: true,
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
		expected: Export("/constants", Define(
			Const("a", "foo"),
			Const("b", 42),
			Const("c", 3.14),
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
		expected: Export("enriched-constants", Define(
			Const("a", "foo"),
			Const("b", 42),
			Const("c", 3.14),
			Query(Rule("url", "https://api.example.org")),
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
		expected: Export("/foo-bar-baz", Define(
			Query(Rule("url", "https://api.example.org")),
			String("foo"),
			Rule("renameField", "foo", "bar"),
			Rule("renameField", "bar", "baz"),
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
		expected: Export("/foo-bar-baz", Define(
			Query(Rule("url", "https://api.example.org")),
			String("foo"),
			Rule("renameField", "foo", "bar"),
			Rule("renameField", "bar", "baz"),
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
		expected: Export("/authenticated-user", Define(
			Query(Rule("url", "https://auth.example.org/info")),
			Query(Rule("authConnector.extended")),
			String("name"),
			Int("level"),
			Float("iris-radius-when-seen-this"),
			Contains("roles", Define(
				Query(
					Rule("url", "https://auth.example.org/roles"),
					Rule("path", Rule("link", "id")),
				),
				String("name"),
				Rule("selectField", "name"),
			)),
		)),
	}}{
		t.Run(test.title, func(t *testing.T) {
			if test.noneExpected {
				testParse(t, test.doc)
			} else if len(test.expectedMany) > 0 {
				testParse(t, test.doc, test.expectedMany...)
			} else {
				testParse(t, test.doc, test.expected)
			}
		})
	}
}
