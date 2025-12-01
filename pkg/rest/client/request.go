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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
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
	nCtx       contextx.IContext

	// enableLogBody was used to record some important info for debug.
	enableLogBody bool

	// enableLogResponse was used to record some important info for debug.
	enableLogResponse bool

	// headerMasker will be used to mask header value.
	headerMasker map[string]func(string) string

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
func (r *Request) WithContext(nCtx contextx.IContext) *Request {
	r.nCtx = nCtx

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
			return nil, fmt.Errorf("failed to marshal body: %v", err)
		}

		return jsonBytes, nil

	case reflect.Struct:
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal struct: %v", err)
		}

		return jsonBytes, nil

	default:
		return nil, fmt.Errorf("unsupported body type, type(%v)", kind)
	}
}

// fullURL get http complete url from request.
func (r *Request) fullURL(endpoint string) *url.URL {
	finalURL, err := url.Parse(endpoint)
	if err != nil {
		r.err = err
		return new(url.URL)
	}

	if len(r.baseURL) != 0 {
		u, err := url.Parse(endpoint + r.baseURL)
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
	logger.G.Biz(r.nCtx).
		WithDuration(time.Since(*start)).
		With("method", r.verb, "url", url, "header", r.maskHeader(r.headers), "body", r.maskRequestBody()).
		Info("http request exceeded max latency time")
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
	FullURL    string
	Body       io.ReadCloser
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

	if r.Body == nil {
		return fmt.Errorf("response body is nil")
	}

	bodyData, err := io.ReadAll(r.Body)
	_ = r.Body.Close()

	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if len(bodyData) == 0 {
		return nil
	}

	logger.G.Sys().With("body", r.maskResponseBody(bodyData), "url", r.FullURL).Info("get response data")

	if r.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("http request failed, status(%d), body(%s)", r.StatusCode, bodyData)
	}

	err = json.Unmarshal(bodyData, obj)
	if nil != err {
		return fmt.Errorf("invalid response body, body(%s): %v", bodyData, err)
	}

	return nil
}

// RawData get raw data.
func (r *Result) RawData() ([]byte, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	if r.Body == nil {
		return nil, fmt.Errorf("response body is nil")
	}

	bodyData, err := io.ReadAll(r.Body)
	_ = r.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if len(bodyData) == 0 {
		return nil, nil
	}

	if r.StatusCode >= http.StatusInternalServerError {
		return nil, fmt.Errorf("http request failed, status(%d), body(%s)", r.StatusCode, bodyData)
	}

	return bodyData, nil
}

// RawStream get raw stream.
func (r *Result) RawStream() (io.ReadCloser, error) {
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
		logger.G.Biz(r.nCtx).
			WithDuration(latency).
			With("method", r.verb, "url", url).
			Warn("throttling request")
	}
}

// Do http request do.
// nolint: nonamedreturns
func (r *Request) Do() (result *Result) {
	if r.err != nil {
		return &Result{
			Err: r.err,
		}
	}

	httpClient := r.capability.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	endpoints, err := r.capability.Discover.GetEndpoints()
	if err != nil {
		return &Result{
			Err: err,
		}
	}

	// tracing
	tracer := r.capability.TraceSvc.TracerProvider().Tracer(r.capability.Name)
	traceCtx, span := tracer.Start(r.nCtx, fmt.Sprintf("%s %s", r.verb, fmt.Sprintf(r.subPath, r.subPathArgs...)),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String(attributeHttpRequestBaseURL, r.baseURL),
			attribute.String(attributeHttpRequestBoby, r.maskRequestBody()),
			attribute.String(attributeHttpRequestHeader, r.maskHeader(r.headers)),
		),
	)
	defer func() {
		span.SetAttributes(
			attribute.Int("http.response.status_code", result.StatusCode),
		)

		span.End()
	}()

	for try := 0; try < r.client.maxRetryCycle; try++ {
		for index, endpoint := range endpoints {
			fullURL := r.fullURL(endpoint).String()
			req, err := r.getRequest(fullURL)
			if err != nil {
				return &Result{Err: err}
			}

			// inject trace context.
			req = req.WithContext(traceCtx)
			r.capability.TraceSvc.TracerPropagator().Inject(traceCtx, propagation.HeaderCarrier(req.Header))

			var isComplete bool
			result, isComplete = r.doWithEndpoint(httpClient, req, try+index)
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

// doWithEndpoint http request do with specific host.
func (r *Request) doWithEndpoint(client HTTPClient, req *http.Request, retries int) (*Result, bool) {
	if retries > 0 {
		r.tryThrottle(req.URL.String())
	}

	logger.G.Biz(r.nCtx).
		With("method", req.Method, "url", req.URL, "header", r.maskHeader(req.Header), "body", r.maskRequestBody()).
		Info("do request")

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		// "Connection reset by peer" is a special err which in most scenario is a transient error.
		// Which means that we can retry it. And so does the VerbTypeGET operation.
		// While the other "write" operation can not simply retry it again, because they are not idempotent.
		r.checkToleranceLatency(&start, req.URL.String())
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
	r.checkToleranceLatency(&start, req.URL.String())

	result := &Result{
		FullURL:           req.URL.String(),
		Body:              resp.Body,
		StatusCode:        resp.StatusCode,
		Status:            resp.Status,
		Header:            resp.Header,
		enableLogResponse: r.enableLogResponse,
	}

	logger.G.Biz(r.nCtx).With(
		"code", result.StatusCode,
		"method", req.Method,
		"url", req.URL.String(),
		"header",
		r.maskHeader(req.Header)).
		Info("receive response")

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

	if r.nCtx != nil {
		req = req.WithContext(r.nCtx)
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

// maskHeader defaultHeaderMasker the http header key.
// nolint: mnd
func (r *Request) maskHeader(headers http.Header) string {
	masked := make(http.Header, len(headers))

	for key, values := range headers {
		maskedValues := make([]string, len(values))
		for i, value := range values {
			if headerMasker, ok := r.headerMasker[key]; ok {
				maskedValues[i] = headerMasker(value)
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

// maskRequestBody defaultHeaderMasker the http body.
// notice: please make sure the request body is necessary and hasn't security risk.
func (r *Request) maskRequestBody() string {
	if !r.enableLogBody {
		return "hidden"
	}

	return string(r.body)
}

// maskResponseBody defaultHeaderMasker the http response body.
// notice: please make sure the response body is necessary and hasn't security risk.
func (r *Result) maskResponseBody(bodyData []byte) string {
	if !r.enableLogResponse {
		return "hidden"
	}

	return string(bodyData)
}
