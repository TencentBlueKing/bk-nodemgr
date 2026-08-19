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

package sshx

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ReadFile reads a regular file from a Unix SSH host. A missing file returns
// empty content so callers can poll files that have not been created yet.
func (cli *Client) ReadFile(ctx context.Context, filePath string) (string, error) {
	if ctx == nil {
		return "", errors.New("context is nil")
	}
	if strings.TrimSpace(filePath) == "" {
		return "", errors.New("file path is empty")
	}

	quotedPath := shellQuote(filePath)
	command := fmt.Sprintf(`if [ -f %s ]; then cat -- %s; fi`, quotedPath, quotedPath)
	content, err := cli.runFileCommand(ctx, command)
	if err != nil {
		return "", fmt.Errorf("read remote file %q: %w", filePath, err)
	}

	return content, nil
}

// ReadLatestFile reads the newest regular file matching a glob pattern from a
// Unix SSH host. No match returns empty content.
func (cli *Client) ReadLatestFile(ctx context.Context, pattern string) (string, error) {
	if ctx == nil {
		return "", errors.New("context is nil")
	}
	if strings.TrimSpace(pattern) == "" {
		return "", errors.New("file pattern is empty")
	}

	command := fmt.Sprintf(
		`file=$(ls -1t -- %s 2>/dev/null | head -n 1); if [ -n "$file" ] && [ -f "$file" ]; then cat -- "$file"; fi`,
		shellQuotePattern(pattern),
	)
	content, err := cli.runFileCommand(ctx, command)
	if err != nil {
		return "", fmt.Errorf("read latest remote file matching %q: %w", pattern, err)
	}

	return content, nil
}

type fileCommandResult struct {
	stdout string
	err    error
}

func (cli *Client) runFileCommand(ctx context.Context, command string) (string, error) {
	resultCh := make(chan fileCommandResult, 1)
	go func() {
		stdout, _, err := cli.RunCommand(command)
		resultCh <- fileCommandResult{stdout: stdout, err: err}
	}()

	select {
	case result := <-resultCh:
		return result.stdout, result.err
	case <-ctx.Done():
		_ = cli.Close()

		return "", fmt.Errorf("context canceled while reading remote file: %w", ctx.Err())
	}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func shellQuotePattern(pattern string) string {
	var builder strings.Builder
	builder.WriteByte('\'')
	for _, char := range pattern {
		switch char {
		case '*', '?', '[', ']':
			builder.WriteByte('\'')
			builder.WriteRune(char)
			builder.WriteByte('\'')
		case '\'':
			builder.WriteString("'\\''")
		default:
			builder.WriteRune(char)
		}
	}
	builder.WriteByte('\'')

	return builder.String()
}
