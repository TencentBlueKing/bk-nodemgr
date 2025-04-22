/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
)

// VerbType http request verb type.
type VerbType string

// http request method.
const (
	// VerbTypePUT ...
	VerbTypePUT VerbType = http.MethodPut
	// VerbTypePOST ...
	VerbTypePOST VerbType = http.MethodPost
	// VerbTypeGET ...
	VerbTypeGET VerbType = http.MethodGet
	// VerbTypeDELETE ...
	VerbTypeDELETE VerbType = http.MethodDelete
	// VerbTypePATCH ...
	VerbTypePATCH VerbType = http.MethodPatch
	// VerbTypeHEAD ...
	VerbTypeHEAD VerbType = http.MethodHead
)

// Request http request.
type Request struct {
	// http client.
	client *Client

	// request capability.
	capability *client.Capability

	verb    VerbType
	params  url.Values
	headers http.Header
	body    []byte
	ctx     context.Context

	// prefixed url
	baseURL string
	// sub path of the url, will be appended to baseURL
	subPath string
	// sub path format args
	subPathArgs []interface{}

	// request timeout value
	timeout time.Duration

	// contentType http content type
	contentType header.ContentType

	err error
}

// WithParams add params to request.
func (r *Request) WithParams(params map[string]string) *Request {
	if r.params == nil {
		r.params = make(url.Values)
	}

	for paramName, value := range params {
		r.params[paramName] = append(r.params[paramName], value)
	}

	return r
}

// WithParam add param to request.
func (r *Request) WithParam(paramName, value string) *Request {
	if r.params == nil {
		r.params = make(url.Values)
	}

	r.params[paramName] = append(r.params[paramName], value)

	return r
}

// WithParamsFromURL add params to request from url.
func (r *Request) WithParamsFromURL(u *url.URL) *Request {
	if r.params == nil {
		r.params = make(url.Values)
	}

	params := u.Query()
	for paramName, value := range params {
		r.params[paramName] = append(r.params[paramName], value...)
	}

	return r
}

// WithHeaders add header to request.
func (r *Request) WithHeaders(header http.Header) *Request {
	if r.headers == nil {
		r.headers = header

		return r
	}

	for key, values := range header {
		for _, v := range values {
			r.headers.Add(key, v)
		}
	}

	return r
}

// WithContext add context to request.
func (r *Request) WithContext(ctx context.Context) *Request {
	r.ctx = ctx

	return r
}

// WithTimeout add timeout to request.
func (r *Request) WithTimeout(d time.Duration) *Request {
	r.timeout = d

	return r
}

// SubResourcef add subPath and subPath's args to request.
func (r *Request) SubResourcef(subPath string, args ...interface{}) *Request {
	r.subPathArgs = args

	return r.subResource(subPath)
}

// subResource add subPath to request.
func (r *Request) subResource(subPath string) *Request {
	subPath = strings.TrimLeft(subPath, "/")
	r.subPath = subPath

	return r
}

// WithContentType add content type to request.
func (r *Request) WithContentType(contentType header.ContentType) *Request {
	r.contentType = contentType

	return r
}

// RawBody add raw body to request.
func (r *Request) RawBody(body []byte) *Request {
	r.body = body

	return r
}

// Body add body to request.
func (r *Request) Body(body interface{}) *Request {
	if body == nil {
		r.body = []byte("")

		return r
	}

	r.body, r.err = processBody(body)
	if r.err != nil {
		return r
	}

	return r
}

// processBody process body.
func processBody(body interface{}) ([]byte, error) {
	valueOf := reflect.ValueOf(body)
	kind := valueOf.Kind()
	switch kind {
	case reflect.String:
		return []byte(valueOf.String()), nil

	case reflect.Ptr:
		if valueOf.IsNil() {
			return []byte(""), nil
		}

		for valueOf.Kind() == reflect.Ptr {
			valueOf = valueOf.Elem()
		}

		return processBody(valueOf.Interface())
	case reflect.Interface, reflect.Slice, reflect.Map:
		if valueOf.IsNil() {
			return []byte(""), nil
		}

		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body, err: %v", err)
		}

		return jsonBytes, nil

	case reflect.Struct:
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal struct, err: %v", err)
		}

		return jsonBytes, nil

	default:
		return nil, fmt.Errorf("unsupported body type, type(%v)", kind)
	}
}

// FullURL get http complete url from request.
func (r *Request) FullURL() *url.URL {
	finalURL := &url.URL{}
	if len(r.baseURL) != 0 {
		u, err := url.Parse(r.baseURL)
		if err != nil {
			r.err = err
			return new(url.URL)
		}
		*finalURL = *u
	}

	if len(r.subPathArgs) > 0 {
		finalURL.Path = finalURL.Path + fmt.Sprintf(r.subPath, r.subPathArgs...)
	} else {
		finalURL.Path = finalURL.Path + r.subPath
	}

	query := url.Values{}
	for key, values := range r.params {
		for _, value := range values {
			query.Add(key, value)
		}
	}

	if r.timeout != 0 {
		query.Set("timeout", r.timeout.String())
	}

	finalURL.RawQuery = query.Encode()

	return finalURL
}

// checkToleranceLatency check request toleranceLatency.
func (r *Request) checkToleranceLatency(start *time.Time, url string, rid string) {
	if time.Since(*start) < r.capability.ToleranceLatencyTime {
		return
	}

	if r.isToleranceLatencyExclusionURL(url) {
		return
	}

	// request time larger than the maxToleranceLatencyTime time, then log the request
	r.capability.Logger.Infof("http request exceeded max latency time. "+
		"cost(%d ms), appcode(%s), user(%s), method(%s), url(%s), body(%s), rid(%s)",
		time.Since(*start)/time.Millisecond,
		r.headers.Get(header.BKAppCodeKey),
		r.headers.Get(header.BKUserKey), r.verb, url, r.body, rid)
}

// isToleranceLatencyExclusionURL judge url if need to checkToleranceLatency.
func (r *Request) isToleranceLatencyExclusionURL(url string) bool {
	for _, eurl := range r.client.exclusionURL {
		if strings.Contains(url, eurl) {
			return true
		}
	}

	return false
}

// Result http response result.
type Result struct {
	Rid        string
	Body       []byte
	Err        error
	StatusCode int
	Status     string
	Header     http.Header
}

// Into parse body to obj.
func (r *Result) Into(obj interface{}) error {
	if r.Err != nil {
		return r.Err
	}

	if len(r.Body) == 0 {
		return nil
	}

	if r.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("http request failed, status(%d), body(%s)", r.StatusCode, r.Body)
	}

	err := json.Unmarshal(r.Body, obj)
	if nil != err {
		if r.StatusCode >= http.StatusMultipleChoices {
			return fmt.Errorf("http request err: %s", string(r.Body))
		}

		return fmt.Errorf("invalid response body, reply(%s), err: %v", r.Body, err.Error())
	}

	return nil
}

// RawData get raw data.
func (r *Result) RawData() ([]byte, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	if r.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("http request failed, status(%d), body(%s)", r.StatusCode, r.Body)
	}

	return r.Body, nil
}

// maxLatency max latency time.
const maxLatency = 200 * time.Millisecond

// tryThrottle try throttle.
func (r *Request) tryThrottle(url string) {
	now := time.Now()

	if latency := time.Since(now); latency > maxLatency {
		r.capability.Logger.Infof("Throttling request took %d ms, verb: %s, request: %s", latency, r.verb, url)
	}
}

// Do http request do.
func (r *Request) Do() *Result {
	rid := getRIDFromContext(r.ctx)
	if rid == "" {
		rid = r.headers.Get(header.BKRIDKey)
	}

	if r.err != nil {
		return &Result{
			Err: r.err,
		}
	}

	client := r.capability.Client
	if client == nil {
		client = http.DefaultClient
	}

	hosts, err := r.capability.Discover.GetServers()
	if err != nil {
		return &Result{
			Err: err,
		}
	}

	for try := 0; try < r.client.maxRetryCycle; try++ {
		for index, host := range hosts {
			result, isComplete := r.doWithHost(client, host, try+index, rid)
			if isComplete {
				return result
			}
		}
	}

	return &Result{
		Err: errors.New("unexpected error"),
	}
}

// retryDelay retry delay.
const retryDelay = 20 * time.Millisecond

// doWithHost http request do with specific host.
func (r *Request) doWithHost(client client.HTTPClient, host string, retries int, rid string) (*Result, bool) {
	contentType := r.contentType

	switch r.contentType {
	case header.FormDataContent:
		r.body = []byte(r.params.Encode())
		r.params = url.Values{}
	case header.JsonContent:
	default:
		contentType = header.JsonContent
	}

	url := host + r.FullURL().String()
	req, err := r.getRequest(url, contentType)
	if err != nil {
		return &Result{Err: err, Rid: rid}, true
	}

	if retries > 0 {
		r.tryThrottle(url)
	}

	r.client.capability.Logger.Infof("restful request, method(%s), url(%s), body(%s), rid(%s)",
		r.verb, url, string(r.body), rid)

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		// "Connection reset by peer" is a special err which in most scenario is a transient error.
		// Which means that we can retry it. And so does the VerbTypeGET operation.
		// While the other "write" operation can not simply retry it again, because they are not idempotent.
		r.checkToleranceLatency(&start, url, rid)
		if !isConnectionReset(err) || r.verb != VerbTypeGET {
			return &Result{Err: err, Rid: rid}, true
		}

		// retry now
		time.Sleep(retryDelay)

		return nil, false
	}

	// collect request metrics.
	r.client.metrics.HandleClientMetrics(req, resp, r.subPath, start)

	// record latency if needed
	r.checkToleranceLatency(&start, url, rid)

	var body []byte
	if resp.Body != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				// retry now
				time.Sleep(retryDelay)
				return nil, false
			}
			r.capability.Logger.Errorf("http request %s %s with body %s, err: %v, rid: %s", string(r.verb), url, r.body,
				err, rid)

			return &Result{Err: err, Rid: rid}, true
		}
		body = data
	}

	return &Result{
		Rid:        rid,
		Body:       body,
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Header:     resp.Header,
	}, true
}

func (r *Request) getRequest(url string, contentType header.ContentType) (*http.Request, error) {
	req, err := http.NewRequest(string(r.verb), url, bytes.NewReader(r.body))
	if err != nil {
		return nil, err
	}

	if r.ctx != nil {
		req = req.WithContext(r.ctx)
	}

	req.Header = cloneHeader(r.headers)
	if len(req.Header) == 0 {
		req.Header = make(http.Header)
	}

	req.Header.Del("Accept-Encoding")
	req.Header.Set("Content-Type", string(contentType))
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// isConnectionReset Returns if the given err is "connection reset by peer" error.
func isConnectionReset(err error) bool {
	var urlErr url.Error
	if errors.Is(err, &urlErr) {
		return false
	}

	var opErr net.OpError
	if errors.Is(err, &opErr) {
		return false
	}

	var osErr os.SyscallError
	if errors.Is(err, &osErr) {
		return false
	}

	var errno syscall.Errno
	if errors.Is(err, errno) && errors.Is(errno, syscall.ECONNRESET) {
		return true
	}

	return false
}

// getRIDFromContext get request id from context.
func getRIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	rid := ctx.Value(header.BKRIDKey)
	ridValue, ok := rid.(string)
	if ok == true {
		return ridValue
	}

	return ""
}

func cloneHeader(src http.Header) http.Header {
	tar := http.Header{}
	for key := range src {
		tar.Set(key, src.Get(key))
	}

	return tar
}
