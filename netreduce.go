package netreduce

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	// TODO: copy required parts and review
	"github.com/stoewer/go-strcase"

	"github.com/netreduce/netreduce/define"
	"github.com/netreduce/netreduce/logging"
)

type ConnectorClient struct {
	client  *http.Client
	baseURL string

	Log Logger
}

type ConnectorContext struct {
	Client ConnectorClient
}

type Connector interface {
	Call(ConnectorContext, RequestContext) (interface{}, error)
}

type RequestContext struct {
	path   string
	cancel chan chan struct{}
}

type Config struct {
	keys map[string]string
}

type Registry struct {
	connectorSpecs []interface{}
	connectors     map[reflect.Value]reflect.Value
	clients        map[Connector]ConnectorClient
	endpoints      map[string]define.Definition
}

type Logger interface {
	Error(...interface{})
}

type Server struct {
	*Config
	*Registry
	Log Logger

	mux         http.ServeMux
	initialized bool
}

type singleChildResult struct {
	name     string
	response interface{}
	err      error
}

var (
	ErrNotFound                  = errors.New("not found")
	ErrGateway                   = errors.New("bad gateway")
	ErrInvalidBackendRequest     = errors.New("invalid backend request")
	ErrUnexpectedBackendResponse = errors.New("unexpected backend response")
	errNotImplemented            = errors.New("not implemented")
)

func newRequestContext(r *http.Request) RequestContext {
	cancel := make(chan chan struct{}, 1)
	cancel <- make(chan struct{})
	c := <-cancel
	cancel <- c
	return RequestContext{
		path:   r.URL.Path,
		cancel: cancel,
	}
}

func (ctx RequestContext) Path() string {
	return ctx.path
}

func (ctx RequestContext) Cancel() {
	c := <-ctx.cancel

	select {
	case <-c:
	default:
		close(c)
	}

	ctx.cancel <- c
}

func (ctx RequestContext) Canceled() <-chan struct{} {
	c := <-ctx.cancel
	ctx.cancel <- c
	return c
}

// TODO: what's the right name for path+query?
func (c ConnectorClient) GetJSON(path string) (interface{}, error) {
	rsp, err := c.client.Get(c.baseURL + path)
	if err != nil {
		c.Log.Error(err)
		return nil, ErrGateway
	}

	defer rsp.Body.Close()

	switch {
	case rsp.StatusCode < http.StatusMultipleChoices:
	case rsp.StatusCode == http.StatusNotFound:
		return nil, ErrNotFound
	case rsp.StatusCode < http.StatusInternalServerError:
		c.Log.Error("invalid backend request", rsp.Status)
		return nil, ErrInvalidBackendRequest
	default:
		c.Log.Error("backend error", rsp.Status)
		return nil, ErrGateway
	}

	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		c.Log.Error("backend error", err)
		return nil, ErrGateway
	}

	var d interface{}
	if err := json.Unmarshal(b, &d); err != nil {
		c.Log.Error("unexpected backend response", err)
		// TODO: whose fault is this typically?
		return nil, ErrUnexpectedBackendResponse
	}

	return d, nil
}

func (c *Config) Get(key string) string { return c.keys[key] }
func (c *Config) Set(key, value string) {
	if c.keys == nil {
		c.keys = make(map[string]string)
	}

	c.keys[key] = value
}

func (r *Registry) SetConnector(c ...interface{}) {
	r.connectorSpecs = append(r.connectorSpecs, c...)
}

func (r *Registry) SetRoute(path string, d define.Definition) {
	if r.endpoints == nil {
		r.endpoints = make(map[string]define.Definition)
	}

	r.endpoints[path] = d
}

func (s *Server) handleRoute(path string, d define.Definition) error {
	handler := func(w http.ResponseWriter, r *http.Request) {
		p := newPlan(d, s.Registry)
		response, err := p.execute(newRequestContext(r))

		if err != nil {
			switch err {
			case ErrNotFound:
				w.WriteHeader(http.StatusNotFound)
			case ErrGateway:
				w.WriteHeader(http.StatusBadGateway)
			case ErrInvalidBackendRequest, ErrUnexpectedBackendResponse:
				w.WriteHeader(http.StatusInternalServerError)
			default:
				s.Log.Error(err)
				w.WriteHeader(http.StatusInternalServerError)
			}

			return
		}

		b, err := json.Marshal(response)
		if err != nil {
			s.Log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(b)
	}

	prefix := path
	if strings.HasSuffix(prefix, "/") {
		prefix = prefix[:len(prefix)-1]
	}

	(&s.mux).Handle(path, http.StripPrefix(prefix, http.HandlerFunc(handler)))
	return nil
}

func (s *Server) Init() error {
	s.initialized = true

	if s.Log == nil {
		s.Log = &logging.Log{}
	}

	if s.Registry == nil {
		return nil
	}

	for _, c := range s.Registry.connectorSpecs {
		t := reflect.TypeOf(c)
		cv := reflect.ValueOf(c)
		urlKey := strings.ToUpper(strcase.SnakeCase(t.Name())) + "_URL"
		baseURL := s.Config.Get(urlKey)
		if baseURL == "" {
			noBaseURL := func() error {
				return fmt.Errorf("cannot decide base URL for: %s", t.Name())
			}

			if t.Kind() != reflect.Struct {
				return noBaseURL()
			}

			buv := cv.FieldByName("BaseURL")
			if buv == (reflect.Value{}) || buv.Type().Kind() != reflect.String {
				return noBaseURL()
			}

			baseURL = buv.String()
			if baseURL == "" {
				return noBaseURL()
			}
		}

		cc := ConnectorClient{
			client:  &http.Client{},
			baseURL: baseURL,
			Log:     s.Log,
		}

		if ct, ok := c.(Connector); ok {
			if s.Registry.clients == nil {
				s.Registry.clients = make(map[Connector]ConnectorClient)
			}

			s.Registry.clients[ct] = cc
		}

		fl := cv.NumField()
		for i := 0; i < fl; i++ {
			if cf, ok := cv.Field(i).Interface().(Connector); ok {
				if s.Registry.clients == nil {
					s.Registry.clients = make(map[Connector]ConnectorClient)
				}

				s.Registry.clients[cf] = cc
			}
		}
	}

	for path, definition := range s.Registry.endpoints {
		if err := s.handleRoute(path, definition); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !s.initialized {
		if err := s.Init(); err != nil {
			panic(err)
		}
	}

	(&s.mux).ServeHTTP(w, r)
}
