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

// Package wiki serves the embedded product user Wiki independently of business APIs.
package wiki

import (
	"bytes"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/wiki"
	"github.com/gin-gonic/gin"
)

// Load registers anonymous, read-only access to the embedded Wiki.
func Load(rg *gin.RouterGroup) {
	rg.GET("/wiki/*filepath", serveFile)
	rg.HEAD("/wiki/*filepath", serveFile)
}

func serveFile(ctx *gin.Context) {
	name := strings.TrimPrefix(ctx.Param("filepath"), "/")
	content, err := wiki.ReadFile(name)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.Header("X-Content-Type-Options", "nosniff")
	if path.Ext(name) == ".md" {
		// Keep Markdown as source text; a future reader can render the same URLs.
		ctx.Header("Content-Type", "text/plain; charset=utf-8")
	}
	http.ServeContent(ctx.Writer, ctx.Request, path.Base(name), time.Time{}, bytes.NewReader(content))
}
