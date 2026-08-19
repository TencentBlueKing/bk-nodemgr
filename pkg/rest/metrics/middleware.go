/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package metrics

import (
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Enable enable metrics into prometheus handler.
func (monitor *Monitor) Enable() *Monitor {
	monitor.bloomFilter = newBloomFilter()

	_ = monitor.metricSets.requestTotal.Enable()
	_ = monitor.metricSets.requestUVTotal.Enable()
	_ = monitor.metricSets.requestBody.Enable()
	_ = monitor.metricSets.responseBody.Enable()
	_ = monitor.metricSets.requestDuration.Enable()
	_ = monitor.metricSets.slowRequest.Enable()

	return monitor
}

// RegisterMiddleware is used to add monitor interceptor to gin router
// It can be called multiple times to intercept from multiple gin.IRoutes.
func (monitor *Monitor) RegisterMiddleware(r gin.IRoutes) *Monitor {
	r.Use(monitor.middleware)

	return monitor
}

// monitorMiddleware as gin monitor middleware.
func (monitor *Monitor) middleware(ctx *gin.Context) {
	// some paths should not be reported
	if slices.Contains(monitor.excludePaths, ctx.Request.URL.Path) {
		ctx.Next()

		return
	}
	startTime := time.Now()

	// execute normal process.
	ctx.Next()

	// after request
	monitor.metricHandle(&metricParam{
		request:               ctx.Request,
		requestPath:           ctx.FullPath(),
		responseStatusCode:    ctx.Writer.Status(),
		responseContentLength: int64(ctx.Writer.Size()),
		processDuration:       time.Since(startTime),
		clientIP:              ctx.ClientIP(),
	})
}

// HandleClientMetrics as a client side metrics handler.
func (monitor *Monitor) HandleClientMetrics(req *http.Request, resp *http.Response, subPath string, start time.Time) {
	if req == nil || resp == nil {
		return
	}

	monitor.metricHandle(&metricParam{
		request:               req,
		requestPath:           subPath,
		responseStatusCode:    resp.StatusCode,
		responseContentLength: resp.ContentLength,
		processDuration:       time.Since(start),
	})
}

type metricParam struct {
	request *http.Request

	// fullpath when handling server metrics.
	// subpath when handling client metrics.
	requestPath string

	responseStatusCode    int
	responseContentLength int64

	processDuration time.Duration

	// client side ip, empty when handling client metrics.
	clientIP string
}

// nolint:cyclop
func (monitor *Monitor) metricHandle(param *metricParam) {
	labels := []string{param.requestPath, param.request.Method, strconv.Itoa(param.responseStatusCode)}

	// set request total
	_ = monitor.metricSets.requestTotal.Inc(labels)

	// set uv
	if !monitor.bloomFilter.contains(param.clientIP) {
		monitor.bloomFilter.add(param.clientIP)
		_ = monitor.metricSets.requestUVTotal.Inc(labels)
	}

	// set request body size
	// since r.ContentLength can be negative (in some occasions) guard the operation
	if param.request.ContentLength >= 0 {
		_ = monitor.metricSets.requestBody.Add(labels, float64(param.request.ContentLength))
	}

	// set slow request
	if param.processDuration >= monitor.slowTime {
		_ = monitor.metricSets.slowRequest.Inc(labels)
	}

	// set request duration
	_ = monitor.metricSets.requestDuration.Observe(labels, float64(param.processDuration.Milliseconds()))

	// set response size
	if param.responseContentLength > 0 {
		_ = monitor.metricSets.responseBody.Add(labels, float64(param.responseContentLength))
	}
}
