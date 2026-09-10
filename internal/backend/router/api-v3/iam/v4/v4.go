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

// Package v4 provides the IAM V4 resource callback router.
package v4

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth/v4/provider"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv4"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

const requestIDHeader = "X-Request-Id"

type handler struct {
	rg           *gin.RouterGroup
	queries      provider.IQueryHandler
	iamV4Handler iamv4.IHandler
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{rg: rg.Group("/v4"), queries: capability.AuthProviderV4Handler, iamV4Handler: capability.IAMV4Handler}
}

// Load registers the dedicated IAM V4 callback transport.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	if capability.IAMV4Handler == nil {
		return
	}
	h := newHandler(rg, capability)
	h.rg.Use(h.basicAuthMiddleware())
	h.rg.POST("/resource", h.handleResourceCallback)
}

func (h *handler) basicAuthMiddleware() gin.HandlerFunc {
	tenantIDPattern := regexp.MustCompile(restserver.TenantIDRegexp)
	return func(gCtx *gin.Context) {
		gCtx.Header(requestIDHeader, gCtx.GetHeader(requestIDHeader))
		rCtx, err := restserver.GenRestContext(gCtx)
		if err != nil {
			callbackError(gCtx, contextx.FromContext(gCtx.Request.Context()), http.StatusInternalServerError, err)
			return
		}
		var tenantID string
		switch tenant.GetMode() {
		case tenant.ModeSingle:
			tenantID = tenant.SingleModeTenantID
		case tenant.ModeMultiple:
			tenantID = gCtx.GetHeader(apigwheader.BKGWTenantIDKey)
			if !tenantIDPattern.MatchString(tenantID) {
				callbackError(gCtx, rCtx, http.StatusBadRequest, fmt.Errorf("invalid tenant id"))
				return
			}
		default:
			callbackError(gCtx, rCtx, http.StatusInternalServerError, fmt.Errorf("invalid tenant mode"))
			return
		}
		// Token validation and downstream GenRestContext must use the same tenant.
		rCtx.Data().SetTenantID(tenantID)
		ctx := contextx.From(rCtx, contextx.WithTenantID(tenantID))
		username, password, ok := gCtx.Request.BasicAuth()
		if !ok {
			callbackError(gCtx, ctx, http.StatusUnauthorized, iamv4.ErrInvalidCredentials)
			return
		}
		if err := h.iamV4Handler.IsBasicAuthAllowed(ctx, username, password); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, iamv4.ErrInvalidCredentials) {
				status = http.StatusUnauthorized
			}
			callbackError(gCtx, ctx, status, err)

			return
		}
		gCtx.Next()
	}
}

func callbackError(gCtx *gin.Context, ctx contextx.IContext, status int, err error) {
	code, message := "INTERNAL", "resource query failed"
	switch status {
	case http.StatusBadRequest:
		code, message = "INVALID_ARGUMENT", "invalid callback arguments"
	case http.StatusUnauthorized:
		code, message = "UNAUTHENTICATED", "invalid callback credentials"
		gCtx.Header("WWW-Authenticate", `Basic realm="IAM"`)
	case http.StatusNotFound:
		code, message = "NOT_FOUND", "resource type or method not found"
	}
	logger.G.Biz(ctx).WithErr(err).With("request-id", gCtx.GetHeader(requestIDHeader)).Error("IAM V4 callback failed")
	gCtx.Header(requestIDHeader, gCtx.GetHeader(requestIDHeader))
	gCtx.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

func (h *handler) handleResourceCallback(gCtx *gin.Context) {
	rCtx, err := restserver.GenRestContext(gCtx)
	if err != nil {
		callbackError(gCtx, contextx.FromContext(gCtx.Request.Context()), http.StatusInternalServerError, err)
		return
	}
	req, err := parseCallbackRequest(gCtx.Request.Body)
	if err != nil {
		callbackError(gCtx, rCtx, http.StatusBadRequest, err)
		return
	}
	result, err := h.query(rCtx, req)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, provider.ErrInvalidArgument):
			status = http.StatusBadRequest
		case errors.Is(err, provider.ErrNotFound):
			status = http.StatusNotFound
		}
		callbackError(gCtx, rCtx, status, err)

		return
	}
	// Encode before writing headers so serialization failures retain the V4 error envelope.
	body, err := json.Marshal(gin.H{"data": result})
	if err != nil {
		callbackError(gCtx, rCtx, http.StatusInternalServerError, err)
		return
	}
	logger.G.Biz(rCtx).With("request-id", gCtx.GetHeader(requestIDHeader)).Debug("IAM V4 callback completed")
	gCtx.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

type callbackRequest struct {
	Type     string          `json:"type"`
	Method   string          `json:"method"`
	Filter   json.RawMessage `json:"filter"`
	Page     json.RawMessage `json:"page"`
	Requires json.RawMessage `json:"requires"`
}

func parseCallbackRequest(body io.Reader) (*callbackRequest, error) {
	decoder := json.NewDecoder(body)
	var req *callbackRequest
	if err := decoder.Decode(&req); err != nil {
		return nil, fmt.Errorf("invalid callback JSON: %w", err)
	}
	if req == nil || strings.TrimSpace(req.Type) == "" || strings.TrimSpace(req.Method) == "" {
		return nil, provider.ErrInvalidArgument
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, provider.ErrInvalidArgument
	}
	for _, field := range []json.RawMessage{req.Filter, req.Page} {
		if len(field) > 0 && field[0] != '{' {
			return nil, provider.ErrInvalidArgument
		}
	}
	if len(req.Requires) > 0 {
		if _, err := parseStringArray(req.Requires); err != nil {
			return nil, err
		}
	}

	return req, nil
}

func (h *handler) query(ctx contextx.IContext, req *callbackRequest) (interface{}, error) {
	if h.queries == nil {
		return nil, fmt.Errorf("resource query handler is nil")
	}
	switch req.Method {
	case "list_instance":
		page, err := parseCallbackPage(req.Page)
		if err != nil {
			return nil, err
		}
		filter, err := parseListFilter(req.Filter)
		if err != nil {
			return nil, err
		}

		return h.queries.ListInstance(ctx, req.Type, &provider.Request[provider.ListInstanceFilter]{Filter: filter, Page: page})
	case "fetch_instance_info":
		var filter struct {
			IDs json.RawMessage `json:"ids"`
		}
		if err := json.Unmarshal(req.Filter, &filter); err != nil {
			return nil, provider.ErrInvalidArgument
		}
		ids, err := parseStringArray(filter.IDs)
		if err != nil {
			return nil, err
		}
		var requires []string
		if len(req.Requires) > 0 {
			requires, err = parseStringArray(req.Requires)
			if err != nil {
				return nil, err
			}
		}

		return h.queries.FetchInstanceInfo(ctx, req.Type, &provider.Request[provider.FetchInstanceFilter]{
			Filter: provider.FetchInstanceFilter{IDs: ids}, Requires: requires,
		})
	default:
		return nil, provider.ErrNotFound
	}
}

func parseStringArray(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 || raw[0] != '[' {
		return nil, provider.ErrInvalidArgument
	}
	var values []*string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, provider.ErrInvalidArgument
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == nil {
			return nil, provider.ErrInvalidArgument
		}
		result = append(result, *value)
	}

	return result, nil
}

func parseCallbackPage(raw json.RawMessage) (types.Page, error) {
	var page struct {
		Page     int64 `json:"page"`
		PageSize int64 `json:"page_size"`
	}
	if err := json.Unmarshal(raw, &page); err != nil {
		return types.Page{}, provider.ErrInvalidArgument
	}
	if page.Page < 1 || page.PageSize < 1 || page.PageSize > provider.MaxListInstancePageSize {
		return types.Page{}, provider.ErrInvalidArgument
	}
	if page.Page-1 > int64(math.MaxInt)/page.PageSize {
		return types.Page{}, provider.ErrInvalidArgument
	}

	return types.Page{Offset: int((page.Page - 1) * page.PageSize), Limit: int(page.PageSize)}, nil
}

func parseListFilter(raw json.RawMessage) (provider.ListInstanceFilter, error) {
	filter := provider.ListInstanceFilter{}
	if len(raw) == 0 {
		return filter, nil
	}
	var fields struct {
		Parent  json.RawMessage `json:"parent"`
		Keyword json.RawMessage `json:"keyword"`
	}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return filter, provider.ErrInvalidArgument
	}
	if len(fields.Keyword) > 0 {
		if fields.Keyword[0] != '"' {
			return filter, provider.ErrInvalidArgument
		}
		if err := json.Unmarshal(fields.Keyword, &filter.Keyword); err != nil {
			return filter, provider.ErrInvalidArgument
		}
	}
	if len(fields.Parent) > 0 {
		if fields.Parent[0] != '{' {
			return filter, provider.ErrInvalidArgument
		}
		if err := json.Unmarshal(fields.Parent, &filter.Parent); err != nil {
			return filter, provider.ErrInvalidArgument
		}
		if filter.Parent.Type == "" || filter.Parent.ID == "" {
			return filter, provider.ErrInvalidArgument
		}
	}

	return filter, nil
}
