package netreduce

import (
	"net/http/httptest"
	"testing"

	"github.com/netreduce/netreduce/data"
	. "github.com/netreduce/netreduce/define"
)

func testWithTestConnector(t *testing.T, route string, d Definition, requestPath string, expected data.Struct) {
	backend := httptest.NewServer(testAPI)
	defer backend.Close()

	config := &Config{}
	config.Set("TEST_CONNECTOR_URL", backend.URL)

	registry := &Registry{}
	registry.SetConnector(testConnector{})
	registry.SetRoute(route, d)

	server := httptest.NewServer(&Server{
		Config:   config,
		Registry: registry,
	})
	defer server.Close()

	result, err := GetJSON(server.URL + requestPath)
	if err != nil {
		t.Fatal(err)
	}

	if !Equal(result, expected) {
		t.Error("invalid response")
		t.Log("got:     ", result)
		t.Log("expected:", expected)
	}
}

func TestConstants(t *testing.T) {
	def := Define(
		Constant("a", "foo"),
		Constant("b", 42),
	)

	expected := data.Struct{
		"a": "foo",
		"b": 42,
	}

	registry := &Registry{}
	registry.SetRoute("/constants", def)

	server := httptest.NewServer(&Server{Registry: registry})
	defer server.Close()

	d, err := GetJSON(server.URL + "/constants")
	if err != nil {
		t.Fatal(err)
	}

	if !Equal(d, expected) {
		t.Error("invalid response")
		t.Log("got:     ", d)
		t.Log("expected:", expected)
	}
}

func TestFixed(t *testing.T) {
	def := Define(
		String("name"),
		Int("age"),
		Query(testConnector{}.GetFixedPerson),
	)

	expected := data.Struct{
		"id":   "fixed-person",
		"name": "John Doe",
		"age":  27,
	}

	testWithTestConnector(t, "/fixed-person", def, "/fixed-person", expected)
}

func TestSelectByPath(t *testing.T) {
	def := Define(
		String("name"),
		Int("age"),
		Query(testConnector{}.GetPerson),
	)

	expected := data.Struct{
		"id":   "foo",
		"name": "John Doe",
		"age":  27,
	}

	testWithTestConnector(t, "/person/", def, "/person/foo", expected)
}

func TestContainsOne(t *testing.T) {
	def := Define(
		String("name"),
		Int("age"),
		ContainsOne("pet", Define(
			String("kind"),
			String("name"),
			Query(testConnector{}.GetPersonsPet),
		)),
		Query(testConnector{}.GetPerson),
	)

	expected := data.Struct{
		"id":   "foo",
		"name": "John Doe",
		"age":  27,
		"pet": data.Struct{
			"id":   "foo-dog",
			"kind": "dog",
			"name": "Winston",
		},
	}

	testWithTestConnector(t, "/person/", def, "/person/foo", expected)
}
