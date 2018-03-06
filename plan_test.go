package netreduce

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/netreduce/netreduce/data"
	. "github.com/netreduce/netreduce/define"
	"github.com/netreduce/netreduce/logging/loggingtest"
)

type concurrentAPI struct {
	syncOne, syncTwo, syncBlock chan struct{}
}

type (
	One   struct{}
	Two   struct{}
	Fail  struct{}
	Block struct{}
)

type syncConnector struct {
	One
	Two
	Fail
	Block
}

func newConcurrentAPI() *concurrentAPI {
	return &concurrentAPI{
		syncOne:   make(chan struct{}),
		syncTwo:   make(chan struct{}),
		syncBlock: make(chan struct{}),
	}
}

func (a *concurrentAPI) one(w http.ResponseWriter, _ *http.Request) {
	select {
	case <-a.syncTwo:
		a.syncOne <- struct{}{}
	case a.syncOne <- struct{}{}:
		<-a.syncTwo
	}

	w.Write([]byte(`{"id": "one"}`))
}

func (a *concurrentAPI) two(w http.ResponseWriter, _ *http.Request) {
	select {
	case <-a.syncOne:
		a.syncTwo <- struct{}{}
	case a.syncTwo <- struct{}{}:
		<-a.syncOne
	}

	w.Write([]byte(`{"id": "two"}`))
}

func (a *concurrentAPI) fail(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func (a *concurrentAPI) block(w http.ResponseWriter, _ *http.Request) {
	<-a.syncBlock
	w.Write([]byte(`{"id": "blocking"}`))
}

func (a *concurrentAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/one":
		a.one(w, r)
	case "/two":
		a.two(w, r)
	case "/fail":
		a.fail(w, r)
	case "/block":
		a.block(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (One) Call(ctx ConnectorContext, _ RequestContext) (interface{}, error) {
	return ctx.Client.GetJSON("/one")
}

func (Two) Call(ctx ConnectorContext, _ RequestContext) (interface{}, error) {
	return ctx.Client.GetJSON("/two")
}

func (Fail) Call(ctx ConnectorContext, _ RequestContext) (interface{}, error) {
	return ctx.Client.GetJSON("/fail")
}

func (Block) Call(ctx ConnectorContext, _ RequestContext) (interface{}, error) {
	return ctx.Client.GetJSON("/block")
}

func TestConnectorsConcurrent(t *testing.T) {
	def := Define(
		ContainsOne("one", Define(
			Query(syncConnector{}.One),
		)),
		ContainsOne("two", Define(
			Query(syncConnector{}.Two),
		)),
	)

	expected := data.Struct{
		"one": data.Struct{"id": "one"},
		"two": data.Struct{"id": "two"},
	}

	api := newConcurrentAPI()
	backend := httptest.NewServer(api)
	defer backend.Close()

	config := &Config{}
	config.Set("SYNC_CONNECTOR_URL", backend.URL)

	registry := &Registry{}
	registry.SetConnector(syncConnector{})
	registry.SetRoute("/", def)

	server := httptest.NewServer(&Server{
		Config:   config,
		Registry: registry,
	})
	defer server.Close()

	done := make(chan struct{})
	go func() {
		defer recover()

		select {
		case <-done:
		case <-time.After(120 * time.Millisecond):
			close(api.syncOne)
			close(api.syncTwo)
			t.Error("timeout")
		}
	}()

	result, err := GetJSON(server.URL)
	close(done)
	if err != nil {
		t.Fatal(err)
	}

	if !Equal(result, expected) {
		t.Error("invalid response")
		t.Log("got:     ", result)
		t.Log("expected:", expected)
	}
}

func TestChildConnectorConcurrent(t *testing.T) {
	def := Define(
		ContainsOne("one", Define(
			Query(syncConnector{}.One),
		)),
		Query(syncConnector{}.Two),
	)

	expected := data.Struct{
		"id":  "two",
		"one": data.Struct{"id": "one"},
	}

	api := newConcurrentAPI()
	backend := httptest.NewServer(api)
	defer backend.Close()

	config := &Config{}
	config.Set("SYNC_CONNECTOR_URL", backend.URL)

	registry := &Registry{}
	registry.SetConnector(syncConnector{})
	registry.SetRoute("/", def)

	server := httptest.NewServer(&Server{
		Config:   config,
		Registry: registry,
	})
	defer server.Close()

	done := make(chan struct{})
	go func() {
		defer recover()

		select {
		case <-done:
		case <-time.After(120 * time.Millisecond):
			close(api.syncOne)
			close(api.syncTwo)
			t.Error("timeout")
		}
	}()

	result, err := GetJSON(server.URL)
	close(done)
	if err != nil {
		t.Fatal(err)
	}

	if !Equal(result, expected) {
		t.Error("invalid response")
		t.Log("got:     ", result)
		t.Log("expected:", expected)
	}
}

func TestConcurrentConnectorCanceled(t *testing.T) {
	def := Define(
		ContainsOne("fail", Define(
			Query(syncConnector{}.Fail),
		)),
		Query(syncConnector{}.Block),
	)

	api := newConcurrentAPI()
	backend := httptest.NewServer(api)
	defer backend.Close()

	config := &Config{}
	config.Set("SYNC_CONNECTOR_URL", backend.URL)

	registry := &Registry{}
	registry.SetConnector(syncConnector{})
	registry.SetRoute("/", def)

	server := httptest.NewServer(&Server{
		Config:   config,
		Registry: registry,
		Log:      loggingtest.NopLog{},
	})
	defer server.Close()

	done := make(chan struct{})
	go func() {
		defer recover()

		select {
		case <-done:
		case <-time.After(120 * time.Millisecond):
			close(api.syncBlock)
			t.Error("timeout")
		}
	}()

	_, err := GetJSON(server.URL)
	close(done)
	close(api.syncBlock)
	if err == nil {
		t.Error("failed to fail")
	}
}
