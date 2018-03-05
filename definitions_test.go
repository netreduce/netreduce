package netreduce

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/netreduce/netreduce/data"
	. "github.com/netreduce/netreduce/define"
)

func apiGetFixed(w http.ResponseWriter, _ *http.Request) {
	fixedPerson := data.Struct{
		"id":   "fixed-person",
		"name": "John Doe",
		"age":  27,
	}

	b, err := json.Marshal(fixedPerson)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func apiGetPerson(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/foo" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	person := data.Struct{
		"id":   "foo",
		"name": "John Doe",
		"age":  27,
	}

	b, err := json.Marshal(person)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func apiGetPersonsPet(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/foo" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	dog := data.Struct{
		"id":   "foo-dog",
		"kind": "dog",
		"name": "Winston",
	}

	b, err := json.Marshal(dog)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

var testAPI = &http.ServeMux{}

func init() {
	testAPI.HandleFunc(
		"/fixed",
		apiGetFixed,
	)

	testAPI.Handle(
		"/person/",
		http.StripPrefix("/person", http.HandlerFunc(apiGetPerson)),
	)

	testAPI.Handle(
		"/pets-by-person/",
		http.StripPrefix("/pets-by-person", http.HandlerFunc(apiGetPersonsPet)),
	)
}

type (
	GetFixedPerson struct{}
	GetPerson      struct{}
	GetPersonsPet  struct{}

	testConnector struct {
		GetFixedPerson
		GetPerson
		GetPersonsPet
	}
)

func (GetFixedPerson) Call(ctx ConnectorContext, arg interface{}) (interface{}, error) {
	return ctx.Client.GetJSON("/fixed")
}

func (GetPerson) Call(ctx ConnectorContext, arg interface{}) (interface{}, error) {
	return ctx.Client.GetJSON("/person" + arg.(string))
}

func (GetPersonsPet) Call(ctx ConnectorContext, arg interface{}) (interface{}, error) {
	return ctx.Client.GetJSON("/pets-by-person" + arg.(string))
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
