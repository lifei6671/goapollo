package goapollo

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
)

var (
	ErrorStatusNotOK    = errors.New("http resp code not ok")
	ErrConfigUnmodified = errors.New("Apollo configuration not changed. ")
)

// this is a static check
var _ requester = (*httpRequester)(nil)

type requester interface {
	request(url string) ([]byte, error)
}

type httpRequester struct {
	client *http.Client
}

func newHTTPRequester(client *http.Client) requester {
	return &httpRequester{
		client: client,
	}
}

func (r *httpRequester) request(url string) ([]byte, error) {
	resp, err := r.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return ioutil.ReadAll(resp.Body)
	}

	// Diacard all body if status code is not 200
	_, _ = io.Copy(ioutil.Discard, resp.Body)

	if resp.StatusCode == http.StatusNotModified {
		return nil, ErrConfigUnmodified
	}

	return nil, ErrorStatusNotOK
}
