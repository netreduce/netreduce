package netreduce_test

import (
	"testing"
	"strings"
	"net/http"
	"net/http/httptest"
	"encoding/json"
	"reflect"

	"github.com/sanity-io/litter"
	"github.com/aryszka/netreduce"
	"github.com/aryszka/netreduce/data"
	. "github.com/aryszka/netreduce/define"
)

// user code
type (
	collectionConnector struct {}
	productConnector struct {}
	stockConnector struct {}
)

func (collectionConnector) GetCollectionByID(netreduce.ConnectorContext, interface{}) interface{} { return nil }
func (stockConnector) GetProductStock(netreduce.ConnectorContext, interface{}) interface{} { return nil }
func (productConnector) GetProductsByCollection(netreduce.ConnectorContext, interface{}) data.List { return nil }

func mapName(collection interface{}) string {
	id := data.String(collection, "id")
	name := id[strings.Index(id, ":"):]
	return name
}

var def = Define(
	StringMapped("name", mapName),
	String("title"),
	ContainsByKey("products", Define(
		String("name"),
		Int("price"),
		ContainsOneByField("stockID", "stock", Define(
			Int("quantity"),
			QueryOne(stockConnector.GetProductStock),
		)),
		Query(productConnector.GetProductsByCollection),
	)),
	QueryOne(collectionConnector.GetCollectionByID),
)
// EO user code

var expectedResponse = data.Struct{
	"id": "collection:foo",
	"name": "foo",
	"title": "Foo Collection",
	"products": data.List{
		data.Struct{
			"id": "product:bar",
			"name": "Bar Product",
			"price": 42,
			"stock": data.Struct{
				"quantity": 2,
			},
		},
		data.Struct{
			"id": "product:qux",
			"name": "Qux Product",
			"price": 81,
			"stock": data.Struct{
				"quantity": 3,
			},
		},
	},
}

func collectionAPI(w http.ResponseWriter, _ *http.Request) {
	b, err := json.Marshal(data.Struct{
		"id": "collection:foo",
		"title": "Foo Collection",
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func productAPI(w http.ResponseWriter, _ *http.Request) {
	b, err := json.Marshal(data.List{
		data.Struct{
			"id": "product:bar",
			"name": "Bar Product",
			"price": 42,
			"stockID": "baz",
		},
		data.Struct{
			"id": "product:qux",
			"name": "Qux Product",
			"price": 81,
			"stockID": "quux",
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func stockAPI(w http.ResponseWriter, r *http.Request) {
	var (
		b []byte
		err error
	)

	switch r.URL.Path {
	case "baz":
		b, err = json.Marshal(data.Struct{
			"quantity": 2,
		})
	case "quux":
		b, err = json.Marshal(data.Struct{
			"quantity": 3,
		})
	default:
		// TODO:
		// - handle errors and optional fields
		// - when is it worth to finish the entire query and when not?
		w.WriteHeader(http.StatusNotFound)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func Test(t *testing.T) {
	collectionServer := httptest.NewServer(http.HandlerFunc(collectionAPI))
	productServer := httptest.NewServer(http.HandlerFunc(productAPI))
	stockServer := httptest.NewServer(http.HandlerFunc(stockAPI))
	defer func() {
		type closer interface { Close() }
		for _, c := range []closer{collectionServer, productServer, stockServer} {
			c.Close()
		}
	}()

	config := &netreduce.Config{}
	config.Set("COLLECTION_URL", collectionServer.URL)
	config.Set("PRODUCT_URL", productServer.URL)
	config.Set("STOCK_URL", stockServer.URL)

	registry := &netreduce.Registry{}
	// TODO: this should not be necessary if the connector doesn't provide any custom settings. In that
	// case, a simple function should be enough in the definition
	registry.SetConnector(
		collectionConnector{},
		productConnector{},
		stockConnector{},
	)
	// TODO: use the last path tag as the id
	registry.SetRoute("/product-collection", def)

	server := &netreduce.Server{
		Config: config,
		Registry: registry,
	}
	if err := server.Init(); err != nil {
		t.Fatal(err)
	}

	htserver := httptest.NewServer(server)
	defer htserver.Close()

	data, err := netreduce.GetJSON(htserver.URL + "/product-collectin")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(data, expectedResponse) {
		t.Error("invalid result")
		t.Log("got:     ", litter.Sdump(data))
		t.Log("expected:", litter.Sdump(expectedResponse))
	}

	// TODO: filters on the APIs
}
