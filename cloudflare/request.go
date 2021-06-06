/*
Copyright 2021 Ahmad Nurus S.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/google/go-querystring/query"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type Request struct {
	baseURL *url.URL
	client  *http.Client
	logger  logr.Logger

	verb       string
	pathPrefix string
	resource   string
	resourceID string
	subpath    string

	headers http.Header
	params  url.Values

	// output
	err  error
	body io.Reader
}

func NewRequest(c *client) *Request {
	base := *c.baseURL
	if !strings.HasSuffix(base.Path, "/") {
		base.Path += "/"
	}
	base.RawQuery = ""
	base.Fragment = ""

	logger := c.logger
	if logger == nil {
		logger = zapr.NewLogger(zap.NewNop())
	}

	return &Request{
		client:  c.client,
		baseURL: &base,
		logger:  logger.WithName("http-request"),
	}
}

func (r *Request) Verb(verb string) *Request {
	r.verb = verb
	return r
}

func (r *Request) Header(key string, values ...string) *Request {
	if r.headers == nil {
		r.headers = http.Header{}
	}
	r.headers.Del(key)
	for _, value := range values {
		r.headers.Add(key, value)
	}
	return r
}

func (r *Request) AccountPrefix(account string) *Request {
	return r.PathPrefix(path.Join("accounts", account))
}

func (r *Request) ZonePrefix(zone string) *Request {
	return r.PathPrefix(path.Join("zones", zone))
}

func (r *Request) PathPrefix(prefix string) *Request {
	r.pathPrefix = prefix
	return r
}

func (r *Request) Resource(resource string) *Request {
	r.resource = resource
	return r
}

func (r *Request) ResourceID(id string) *Request {
	r.resourceID = id
	return r
}

func (r *Request) SubPath(path string) *Request {
	r.subpath = path
	return r
}

func (r *Request) Param(params interface{}) *Request {
	if r.params == nil {
		r.params = url.Values{}
	}

	p, err := query.Values(params)
	if err != nil {
		r.err = err
	}

	r.params = p
	return r
}

func (r *Request) Body(body interface{}) *Request {
	switch bt := body.(type) {
	case []byte:
		r.body = bytes.NewReader(bt)
	case io.Reader:
		r.body = bt
	default:
		data, err := json.Marshal(bt)
		if err != nil {
			r.err = err
			return r
		}

		r.body = bytes.NewReader(data)
	}

	r.Header("Content-Type", "application/json")
	return r
}

// URL build url from given r.
// url.Path will constructed as following:
//	/{basePath}/{pathPrefix}/{resourceName}/{resourceID}/{subPath}
//
// where are:
//
// 	- basePath		= `/client/v4`
// 	- pathPrefix		= could be `/accounts/{account_id}` or `/zones/{zone_id}` (optional)
// 	- resourceName		= cloudflare resources endpoint name eg: `/tunnels`
// 	- resourceID		= cloudflare resource id `/:uuid` (optional)
// 	- subPath		= cloudflare sub resource endpoint eg: `/routes`
//
// beside path, this function also construct url.Query from given r.params
func (r *Request) URL() *url.URL {
	baseURL := r.baseURL
	basePath := baseURL.Path
	p := basePath
	if r.pathPrefix != "" {
		p = path.Join(p, r.pathPrefix)
	}

	p = path.Join(p, strings.ToLower(r.resource))
	if r.resourceID != "" {
		p = path.Join(p, r.resourceID)
	}

	if r.subpath != "" {
		p = path.Join(p, r.subpath)
	}

	url := baseURL
	url.Path = p
	url.RawQuery = r.params.Encode()
	return url
}

func (r *Request) Do(ctx context.Context) RequestResult {
	if r.err != nil {
		r.logger.V(1).Error(r.err, "Error before making http request")
		return RequestResult{err: r.err}
	}

	var result RequestResult
	req, err := http.NewRequest(r.verb, r.URL().String(), r.body)
	if err != nil {
		r.logger.V(1).Error(err, "Unable to make http request")
		return RequestResult{err: err}
	}

	req = req.WithContext(ctx)
	req.Header = r.headers
	reqTime := time.Now()
	resp, err := r.client.Do(req)
	if err != nil {
		r.logger.V(1).Error(err, "Failed calling http request")
		return RequestResult{err: err}
	}

	defer resp.Body.Close()
	r.logger.V(1).Info("Calling http request", "method", req.Method, "url", req.URL.String(), "status", resp.Status, "in", time.Since(reqTime))
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		r.logger.V(1).Error(err, "Unable to read response body")
		return RequestResult{err: err, statusCode: resp.StatusCode}
	}

	var res Response
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&res); err != nil {
		r.logger.V(1).Error(err, "Unable to decode response body")
		return RequestResult{err: err, statusCode: resp.StatusCode}
	}

	result.raw = data
	result.err = err
	result.res = res
	result.statusCode = resp.StatusCode
	return result
}

type RequestResult struct {
	raw        []byte
	res        Response
	err        error
	statusCode int
}

func (r RequestResult) Error() error {
	if r.err != nil {
		return r.err
	}

	if len(r.res.Errors) != 0 {
		var err error
		for _, e := range r.res.Errors {
			err = multierr.Append(err, e)
		}

		return err
	}

	return nil
}

func (r RequestResult) Raw() ([]byte, error) {
	return r.raw, r.err
}

func (r RequestResult) StatusCode() int {
	return r.statusCode
}

func (r RequestResult) Into(o interface{}) error {
	if err := r.Error(); err != nil {
		return err
	}

	return json.NewDecoder(bytes.NewReader(r.res.Result)).Decode(o)
}
