package engine

import (
	"bytes"
	"testing"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/netreduce/netreduce/nred"
	"github.com/netreduce/netreduce/data"
	"github.com/netreduce/netreduce/registry"
)

func createDef(format string, args ...interface{}) (d nred.Definition, err error) {
	doc := fmt.Sprintf(format, args...)

	var defs []nred.Definition
	if defs, err = nred.Parse(bytes.NewBufferString(doc)); err != nil {
		return
	}

	if len(defs) != 1 {
		err = fmt.Errorf("invalid count of definitions: %d", len(defs))
		return
	}

	d = defs[0]
	return
}

func execDefRequest(d nred.Definition, req Incoming) (data.Data, error) {
	r := registry.New()
	if err := r.RegisterBuiltinRules(); err != nil {
		return data.Zero(), err
	}

	return New(Options{registry: r}).ExecDefinition(d, req)
}

func execDef(d nred.Definition) (data.Data, error) {
	var req Incoming
	return execDefRequest(d, req)
}

func Test(t *testing.T) {
	t.Run("hello", func(t *testing.T) {
		def, err := createDef(`"Hello, world!"`)
		if err != nil {
			t.Fatal(err)
		}

		d, err := execDef(def)
		if err != nil {
			t.Fatal(err)
		}

		if d.String() != `"Hello, world!"` {
			t.Error("invalid data received")
			t.Log("got:     ", d)
			t.Log("expected:", `"Hello, world!"`)
			return
		}
	})

	t.Run("pass through", func(t *testing.T) {
		b := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte(`"Hello, world!"`))
		}))
		defer b.Close()

		def, err := createDef(`query("%s")`, b.URL)
		if err != nil {
			t.Fatal(err)
		}

		d, err := execDef(def)
		if err != nil {
			t.Fatal(err)
		}

		if d.String() != `"Hello, world!"` {
			t.Error("invalid data received")
			t.Log("got:     ", d)
			t.Log("expected:", `"Hello, world!"`)
			return
		}
	})

	t.Run("empty", func(t *testing.T) {
		var def nred.Definition
		d, err := execDef(def)
		if err != nil {
			t.Fatal(err)
		}

		if !data.IsZero(d) {
			t.Error("invalid data received")
			t.Log("got:     ", d)
			t.Log("expected:", `"Hello, world!"`)
			return
		}
	})

	t.Run("constants", func(t *testing.T) {
		def, err := createDef(`define(
			const("a", "foo")
			const("b", 42)
			const("c", 3.14)
		)`)
		if err != nil {
			t.Fatal(err)
		}

		d, err := execDef(def)
		if err != nil {
			t.Fatal(err)
		}

		expect := data.Struct(map[string]data.Data{
			"a": data.String("foo"),
			"b": data.Int(42),
			"c": data.Float(3.14),
		})
		if !data.Eq(d, expect) {
			t.Error("invalid data received")
			t.Log("got:     ", d)
			t.Log("expected:", expect)
			return
		}
	})

	t.Run("enriched constants", func(t *testing.T) {
		b := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte(`{"id": "foo1", "foo": 1}`))
		}))
		defer b.Close()

		def, err := createDef(`
			let constants = define(
				const("a", "foo")
				const("b", 42)
				const("c", 3.14)
			)

			export "enriched-constants" define(
				constants
				query("%s")
				string("foo")
			)
		`, b.URL)
		if err != nil {
			t.Fatal(err)
		}

		d, err := execDef(def)
		if err != nil {
			t.Fatal(err)
		}

		expect := data.Struct(map[string]data.Data{
			"id": data.String("foo1"),
			"a": data.String("foo"),
			"b": data.Int(42),
			"c": data.Float(3.14),
			"foo": data.Int(1),
		})
		if !data.Eq(d, expect) {
			t.Error("invalid data received")
			t.Log("got:     ", d)
			t.Log("expected:", expect)
			return
		}
	})

	t.Run("user details", func(t *testing.T) {
		users := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/user/user1" {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.Write([]byte(`{"id": "u1", "username": "user1", "name": "User One"}`))
		}))
		defer users.Close()

		userDetails := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/user-details/u1" {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.Write([]byte(`{"id": "u1", "level": 2}`))
		}))
		defer userDetails.Close()

		userRoles := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/u1" {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.Write([]byte(`[
				{"id": 0, "name": "admin"},
				{"id": 1, "name": "user"},
				{"id": 2, "name": "netop"}
			]`))
		}))
		defer userRoles.Close()

		def, err := createDef(
			`define(
				query("%s/user", appendPath(param("username")))
				query("%s/user-details", appendPath(link("id")))

				string("name")
				int("level")

				contains("roles", define(
					query("%s", setPath(link("id")))
					string("name")
					selectField("name")
				))
			)`,
			users.URL,
			userDetails.URL,
			userRoles.URL,
		)
		if err != nil {
			t.Fatal(err)
		}

		d, err := execDefRequest(def, Incoming{
			params: map[string]interface{}{"username": "user1"},
		})
		if err != nil {
			t.Fatal(err)
		}

		expect := data.Struct(map[string]data.Data{
			"id": data.String("u1"),
			"name": data.String("User One"),
			"level": data.Int(2),
			"roles": data.List([]data.Data{
				data.String("admin"),
				data.String("user"),
				data.String("netop"),
			}),
		})
		if !data.Eq(d, expect) {
			t.Error("invalid data received")
			t.Log("got:     ", d)
			t.Log("expected:", expect)
			return
		}
	})
}
