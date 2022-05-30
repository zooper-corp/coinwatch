package tools

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// ReadHTTPRequest reads data from an http request
func ReadHTTPRequest(req *http.Request, client *http.Client) ([]byte, error, int, http.Header) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err, 400, http.Header{}
	}
	return ReadHTTPResponse(resp)
}

// ReadHTTPResponse reads data from an http response
func ReadHTTPResponse(resp *http.Response) ([]byte, error, int, http.Header) {
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(resp.Body)
	default:
		reader = resp.Body
	}
	// Defer close
	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)
	// Read
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err, 400, http.Header{}
	}
	if 200 != resp.StatusCode {
		return nil, fmt.Errorf("%s", body), resp.StatusCode, resp.Header
	}
	return body, nil, 200, resp.Header
}
