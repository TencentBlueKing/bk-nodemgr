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

// Package jinja2x provides jinja2x render.
package jinja2x

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tmp"
)

// Handler the handler for jinja2x.
type Handler struct {
}

// Render implement IHandler.
func (h Handler) Render(tmpl string, content map[string]any) (string, error) {
	templateBuffer := strings.NewReader(tmpl)
	templateFile, err := tmp.NewTempFileWithSpecialName(io.NopCloser(templateBuffer), "template.conf")
	if err != nil {
		return "", fmt.Errorf("failed to create template file: %w", err)
	}

	defer func() {
		cleanErr := templateFile.CleanUp()
		if cleanErr != nil {
			logger.G.Sys().WithErr(cleanErr).Warn("failed to clean up temp file")
		}
	}()

	contentByes, err := json.Marshal(content)
	if err != nil {
		return "", fmt.Errorf("failed to marshal content: %w", err)
	}

	contentBuffer := bytes.NewReader(contentByes)
	contentFile, err := tmp.NewTempFileWithSpecialName(io.NopCloser(contentBuffer), "content.json")
	if err != nil {
		return "", fmt.Errorf("failed to create content file: %w", err)
	}

	defer func() {
		cleanErr := contentFile.CleanUp()
		if cleanErr != nil {
			logger.G.Sys().WithErr(cleanErr).Warn("failed to clean up temp file")
		}
	}()

	resultFile, err := tmp.NewTempFileWithSpecialName(io.NopCloser(strings.NewReader("")), "result.conf")
	if err != nil {
		return "", fmt.Errorf("failed to create result file: %w", err)
	}

	defer func() {
		cleanErr := resultFile.CleanUp()
		if cleanErr != nil {
			logger.G.Sys().WithErr(cleanErr).Warn("failed to clean up temp file")
		}
	}()

	args := []string{
		"-t",
		fmt.Sprintf("%s", templateFile.Path()),
		"-c",
		fmt.Sprintf("%s", contentFile.Path()),
		"-o",
		fmt.Sprintf("%s", resultFile.Path()),
	}
	stdout, stderr, err := jinja2ExecRunCmd(context.Background(), args, os.Environ())
	if err != nil {
		return "", fmt.Errorf("failed to render template, stdout(%s), stderr(%s): %w", stdout, stderr, err)
	}

	result, err := os.ReadFile(resultFile.Path())
	if err != nil {
		return "", fmt.Errorf("failed to read result file: %w", err)
	}

	return string(result), nil
}

// New create a new Handler.
func New() *Handler {
	return &Handler{}
}

// nolint: gochecknoglobals
var handler = New()

// Render render the template with context.
func Render(tmpl string, context map[string]any) (string, error) {
	return handler.Render(tmpl, context)
}
