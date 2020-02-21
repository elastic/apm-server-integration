package tests

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
)

type recorder struct {
	url string
	resp *http.Response
}

func (c *recorder) sendRequest(r io.Reader, h http.Header) error {
	var req *http.Request
	var err error
	if r == nil {
		req, err = http.NewRequest(http.MethodGet, c.url, nil)
	} else {
		req, err = http.NewRequest(http.MethodPost, c.url, r)
	}
	if err != nil {
		return err
	}
	req.Header = h
	c.resp, err = http.DefaultClient.Do(req)
	return err
}

func (c *recorder) response() (int, map[string]interface{}, error) {
	if c.resp == nil {
		return 0, nil, errors.New("screwed something up")
	}
	bytes, err:= ioutil.ReadAll(c.resp.Body)
	if err != nil {
		return c.resp.StatusCode, nil, err
	}
	var body map[string]interface{}
	if len(bytes) > 0 {
		err = json.Unmarshal(bytes, &body)
	}
	return c.resp.StatusCode, body, err
}
