package nrhttp

import (
	"net/http"
	"io/ioutil"

	"github.com/netreduce/netreduce/data"
)

func Get(url string) (data.Data, error) {
	rsp, err := http.Get(url)
	if err != nil {
		return data.Zero(), err
	}

	defer rsp.Body.Close()

	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return data.Zero(), err
	}

	return data.JSON(b), nil
}
