/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package client

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
	capability *Capability

	verb       VerbType
	params     url.Values
	headers    http.Header
	body       []byte
	bodyReader io.Reader
	ctx        context.Context

	// enableLogBody was used to record some important info for debug.
	enableLogBody bool

	// enableLogResponse was used to record some important info for debug.
	enableLogResponse bool

	// sensitive headers.
	sensitiveHeaders map[string]struct{}

	// prefixed url
	baseURL string
	// sub path of the url, will be appended to baseURL
	subPath string
	// sub path format args
	subPathArgs []interface{}

	// request timeout value
	timeout time.Duration

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

// BodyReader add reader to request.
func (r *Request) BodyReader(reader io.Reader) *Request {
	r.bodyReader = reader

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

	// in this case, we could make sure the finalURL.Path and subPath both are valid.
	if len(r.subPathArgs) > 0 {
		finalURL.Path, _ = url.JoinPath(finalURL.Path, fmt.Sprintf(r.subPath, r.subPathArgs...))
	} else {
		finalURL.Path, _ = url.JoinPath(finalURL.Path, r.subPath)
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
func (r *Request) checkToleranceLatency(start *time.Time, url string) {
	if time.Since(*start) < r.capability.ToleranceLatencyTime {
		return
	}

	if r.isToleranceLatencyExclusionURL(url) {
		return
	}

	// request time larger than the maxToleranceLatencyTime time, then log the request
	r.capability.Logger.Infof("http request exceeded max latency time. "+
		"cost(%d ms), method(%s), url(%s), header(%s), body(%s)",
		time.Since(*start)/time.Millisecond, r.verb, url, r.maskHeader(r.headers), r.maskRequestBody())
}

// isToleranceLatencyExclusionURL judge url if need to checkToleranceLatency.
func (r *Request) isToleranceLatencyExclusionURL(url string) bool {
	for eurl := range r.client.exclusionURL {
		if strings.Contains(url, eurl) {
			return true
		}
	}

	return false
}

// Result http response result.
type Result struct {
	Body       []byte
	Err        error
	StatusCode int
	Status     string
	Header     http.Header

	// enableLogResponse was used to record some important info for debug.
	enableLogResponse bool
}

// Into parse body to obj.
func (r *Result) Into(obj interface{}) error {
	if r.Err != nil {
		return r.Err
	}

	if len(r.Body) == 0 {
		return nil
	}

	if r.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("http request failed, status(%d), body(%s)", r.StatusCode, r.Body)
	}

	err := json.Unmarshal(r.Body, obj)
	if nil != err {
		return fmt.Errorf("invalid response body, body(%s): %v", r.Body, err)
	}

	return nil
}

// RawData get raw data.
func (r *Result) RawData() ([]byte, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	if r.StatusCode >= http.StatusInternalServerError {
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
	if r.err != nil {
		return &Result{
			Err: r.err,
		}
	}

	httpClient := r.capability.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	servers, err := r.capability.Discover.GetServers()
	if err != nil {
		return &Result{
			Err: err,
		}
	}

	for try := 0; try < r.client.maxRetryCycle; try++ {
		for index, host := range servers {
			result, isComplete := r.doWithHost(httpClient, host, try+index)
			if isComplete {
				return result
			}
		}
	}

	return &Result{
		Err: errors.New("request unexpected error"),
	}
}

// retryDelay retry delay.
const retryDelay = 20 * time.Millisecond

// doWithHost http request do with specific host.
func (r *Request) doWithHost(client HTTPClient, host string, retries int) (*Result, bool) {
	url := host + r.FullURL().String()
	req, err := r.getRequest(url)
	if err != nil {
		return &Result{Err: err}, true
	}

	if retries > 0 {
		r.tryThrottle(url)
	}

	r.client.capability.Logger.Infof("request, method(%s), url(%s), header(%s), body(%s)",
		r.verb, url, r.maskHeader(r.headers), r.maskRequestBody())

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		// "Connection reset by peer" is a special err which in most scenario is a transient error.
		// Which means that we can retry it. And so does the VerbTypeGET operation.
		// While the other "write" operation can not simply retry it again, because they are not idempotent.
		r.checkToleranceLatency(&start, url)
		if !isConnectionReset(err) || r.verb != VerbTypeGET {
			return &Result{Err: err}, true
		}

		// retry now
		time.Sleep(retryDelay)

		return nil, false
	}

	// collect request metrics.
	r.client.metrics.HandleClientMetrics(req, resp, r.subPath, start)

	// record latency if needed
	r.checkToleranceLatency(&start, url)

	var body []byte
	if resp.Body != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				// retry now
				time.Sleep(retryDelay)
				return nil, false
			}
			r.capability.Logger.Errorf("failed to request, method(%s), url(%s), header(%s), body(%s): %v",
				r.verb, url, r.maskHeader(r.headers), r.maskRequestBody(), err)

			return &Result{Err: err}, true
		}
		body = data
	}

	result := &Result{
		Body:              body,
		StatusCode:        resp.StatusCode,
		Status:            resp.Status,
		Header:            resp.Header,
		enableLogResponse: r.enableLogResponse,
	}

	r.client.capability.Logger.Infof(
		"response, method(%s), url(%s), header(%s), http-code(%d), body(%s)",
		r.verb, url, r.maskHeader(r.headers), result.StatusCode, result.maskResponseBody())

	return result, true
}

func (r *Request) getRequest(url string) (*http.Request, error) {
	reader := r.bodyReader
	if reader == nil {
		reader = bytes.NewReader(r.body)
	}
	req, err := http.NewRequest(string(r.verb), url, reader)
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
	req.Header.Set("Accept", "*/*")
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

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

func cloneHeader(src http.Header) http.Header {
	tar := http.Header{}
	for key := range src {
		tar.Set(key, src.Get(key))
	}

	return tar
}

// maskHeader mask the http header key.
// nolint: mnd
func (r *Request) maskHeader(headers http.Header) string {
	masked := make(http.Header, len(headers))

	for key, values := range headers {
		maskedValues := make([]string, len(values))
		for i, value := range values {
			if _, ok := r.sensitiveHeaders[key]; ok {
				if len(value) > 6 {
					maskedValues[i] = value[:3] + "***" + value[len(value)-3:]
				} else {
					maskedValues[i] = strings.Repeat("*", len(value))
				}
			} else {
				maskedValues[i] = value
			}
		}
		masked[key] = maskedValues
	}

	return fmt.Sprintf("%+v", masked)
}

// EnableLogBody show the request body.
func (r *Request) EnableLogBody() *Request {
	r.enableLogBody = true

	return r
}

// EnableLogResponse show the request response.
func (r *Request) EnableLogResponse() *Request {
	r.enableLogResponse = true

	return r
}

// maskRequestBody mask the http body.
// notice: please make sure the request body is necessary and hasn't security risk.
func (r *Request) maskRequestBody() string {
	if !r.enableLogBody {
		return "hidden"
	}

	return string(r.body)
}

// maskResponseBody mask the http response body.
// notice: please make sure the response body is necessary and hasn't security risk.
func (r *Result) maskResponseBody() string {
	if !r.enableLogResponse {
		return "hidden"
	}

	return string(r.Body)
}
