package netreduce

import (
	"encoding/json"
	"net/http"

	"github.com/netreduce/netreduce/data"
)

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

func (GetFixedPerson) Call(ctx ConnectorContext, _ RequestContext) (interface{}, error) {
	return ctx.Client.GetJSON("/fixed")
}

func (GetPerson) Call(ctx ConnectorContext, rctx RequestContext) (interface{}, error) {
	return ctx.Client.GetJSON("/person" + rctx.Path())
}

func (GetPersonsPet) Call(ctx ConnectorContext, rctx RequestContext) (interface{}, error) {
	return ctx.Client.GetJSON("/pets-by-person" + rctx.Path())
}
