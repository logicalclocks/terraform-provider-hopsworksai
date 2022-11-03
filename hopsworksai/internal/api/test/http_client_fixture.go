package test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type HttpClientFixture struct {
	ExpectMethod       string
	ExpectPath         string
	ExpectHeaders      map[string]string
	ExpectRequestBody  string
	ExpectRequestQuery string
	ResponseBody       string
	ResponseCode       int
	ReturnError        error
	FailWithError      string
	T                  *testing.T
}

func CompactJSONString(jsonString string) string {
	for _, v := range []string{"\n", "\t", " "} {
		jsonString = strings.ReplaceAll(jsonString, v, "")
	}
	return jsonString
}

func (c *HttpClientFixture) Do(req *http.Request) (*http.Response, error) {
	if c.T == nil {
		return nil, fmt.Errorf("missing testing.T in HttpClientFixture")
	}

	if c.FailWithError != "" {
		c.T.Fatal(c.FailWithError)
	}

	if c.ExpectMethod != "" && req.Method != c.ExpectMethod {
		c.T.Fatalf("invalid method, expected %s but got %s", c.ExpectMethod, req.Method)
	}

	if c.ExpectPath != "" && req.URL.Path != c.ExpectPath {
		c.T.Fatalf("invalid path, expected %s but got %s", c.ExpectPath, req.URL.Path)
	}

	if c.ExpectRequestQuery != "" && req.URL.RawQuery != c.ExpectRequestQuery {
		c.T.Fatalf("invalid request query, expected %s but got %s", c.ExpectRequestQuery, req.URL.RawQuery)
	}

	if c.ExpectHeaders != nil {
		for k, v := range c.ExpectHeaders {
			o := req.Header.Get(k)
			if v != o {
				c.T.Fatalf("invalid header, expected %s but got %s", v, o)
			}
		}
	}
	if c.ExpectRequestBody != "" {
		reqBody, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		reqBodyString := string(reqBody)
		expected := CompactJSONString(c.ExpectRequestBody)
		if reqBodyString != expected {
			c.T.Fatalf("invalid req body, expected:\n%s, but got:\n%s", expected, reqBodyString)
		}
	}
	return &http.Response{
		StatusCode: c.ResponseCode,
		Body:       io.NopCloser(strings.NewReader(c.ResponseBody)),
	}, c.ReturnError
}
