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

package systeminfo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
)

const testStep logger.Step = "general"

func TestTargetInfoFieldsUseFixedKeyOrder(t *testing.T) {
	expectedKeys := []string{
		"goos",
		"goarch",
		"hostname",
		"current_time",
		"timezone",
		"timezone_offset",
		"system_version",
		"kernel_version",
	}

	if len(targetInfoFields) != len(expectedKeys) {
		t.Fatalf("targetInfoFields length = %d, want %d", len(targetInfoFields), len(expectedKeys))
	}

	for i, expectedKey := range expectedKeys {
		if targetInfoFields[i].key != expectedKey {
			t.Fatalf("targetInfoFields[%d].key = %q, want %q", i, targetInfoFields[i].key, expectedKey)
		}
	}
}

func TestLogInitialTargetInfoLogsFixedFields(t *testing.T) {
	fields := []targetInfoField{
		newTestField("goos", "linux", nil),
		newTestField("goarch", "amd64", nil),
		newTestField("hostname", "bk-node", nil),
		newTestField("current_time", "2026-07-08T10:11:12+08:00", nil),
		newTestField("timezone", "CST", nil),
		newTestField("timezone_offset", "+08:00", nil),
		newTestField("system_version", "BlueKing Linux", nil),
		newTestField("kernel_version", "6.6.0", nil),
	}

	setTargetInfoFields(t, fields)
	logs := captureInitialTargetInfoLogs(t)

	expectedLogs := []string{
		expectedStartLine(),
		expectedFieldDebugLine("goos", "linux"),
		expectedFieldDebugLine("goarch", "amd64"),
		expectedFieldDebugLine("hostname", "bk-node"),
		expectedFieldDebugLine("current_time", "2026-07-08T10:11:12+08:00"),
		expectedFieldDebugLine("timezone", "CST"),
		expectedFieldDebugLine("timezone_offset", "+08:00"),
		expectedFieldDebugLine("system_version", "BlueKing Linux"),
		expectedFieldDebugLine("kernel_version", "6.6.0"),
	}
	assertEqualStrings(t, expectedLogs, logs)
}

func TestLogInitialTargetInfoUsesUnknownWithoutWarningForEmptyValue(t *testing.T) {
	setTargetInfoFields(t, []targetInfoField{
		newTestField("system_version", "", nil),
	})

	logs := captureInitialTargetInfoLogs(t)

	expectedLogs := []string{
		expectedStartLine(),
		expectedFieldDebugLine("system_version", "unknown"),
	}
	assertEqualStrings(t, expectedLogs, logs)
	assertNoWarnLog(t, logs)
}

func TestLogInitialTargetInfoWarnsBeforeUnknownDebugOnFailure(t *testing.T) {
	collectErr := errors.New("collector failed")
	setTargetInfoFields(t, []targetInfoField{
		newTestField("system_version", "", collectErr),
		newTestField("kernel_version", "6.6.0", nil),
	})

	logs := captureInitialTargetInfoLogs(t)

	expectedLogs := []string{
		expectedStartLine(),
		expectedFieldWarnLine("system_version", collectErr),
		expectedFieldDebugLine("system_version", "unknown"),
		expectedFieldDebugLine("kernel_version", "6.6.0"),
	}
	assertEqualStrings(t, expectedLogs, logs)
}

func TestLogInitialTargetInfoDoesNotPanicWhenCollectorsFail(t *testing.T) {
	collectErr := errors.New("collector failed")
	fields := []targetInfoField{
		newTestField("goos", "", collectErr),
		newTestField("goarch", "", collectErr),
		newTestField("hostname", "", collectErr),
		newTestField("current_time", "", collectErr),
		newTestField("timezone", "", collectErr),
		newTestField("timezone_offset", "", collectErr),
		newTestField("system_version", "", collectErr),
		newTestField("kernel_version", "", collectErr),
	}
	setTargetInfoFields(t, fields)

	logs := captureInitialTargetInfoLogs(t)

	if len(logs) != 1+len(fields)*2 {
		t.Fatalf("log count = %d, want %d", len(logs), 1+len(fields)*2)
	}
	assertEqualStrings(t, []string{expectedStartLine()}, logs[:1])
	for i, field := range fields {
		warnIndex := 1 + i*2
		debugIndex := warnIndex + 1
		assertEqualStrings(t, []string{
			expectedFieldWarnLine(field.key, collectErr),
			expectedFieldDebugLine(field.key, "unknown"),
		}, logs[warnIndex:debugIndex+1])
	}
}

func TestCollectCurrentTimeReturnsRFC3339(t *testing.T) {
	value, err := collectCurrentTime()
	if err != nil {
		t.Fatalf("collectCurrentTime() error = %v, want nil", err)
	}
	if _, err := time.Parse(time.RFC3339, value); err != nil {
		t.Fatalf("collectCurrentTime() = %q, want RFC3339: %v", value, err)
	}
}

func TestFormatTimezoneOffset(t *testing.T) {
	tests := []struct {
		name          string
		offsetSeconds int
		want          string
	}{
		{
			name:          "zero offset",
			offsetSeconds: 0,
			want:          "+00:00",
		},
		{
			name:          "positive hours and minutes",
			offsetSeconds: 8*3600 + 30*60,
			want:          "+08:30",
		},
		{
			name:          "negative hours and minutes",
			offsetSeconds: -(5*3600 + 45*60),
			want:          "-05:45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTimezoneOffset(tt.offsetSeconds)
			if got != tt.want {
				t.Fatalf("formatTimezoneOffset(%d) = %q, want %q", tt.offsetSeconds, got, tt.want)
			}
		})
	}
}

func captureInitialTargetInfoLogs(t *testing.T) []string {
	t.Helper()

	var buf bytes.Buffer
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()
	originalLevel := logger.GetLevel()

	log.SetOutput(&buf)
	log.SetFlags(0)
	log.SetPrefix("")
	if err := logger.SetLevel(logger.LevelDebug); err != nil {
		t.Fatalf("set logger level: %v", err)
	}
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
		log.SetFlags(originalFlags)
		log.SetPrefix(originalPrefix)
		if err := logger.SetLevel(originalLevel); err != nil {
			t.Fatalf("restore logger level: %v", err)
		}
	})

	LogInitialTargetInfo(testStep)
	log.SetOutput(io.Discard)

	return splitLogLines(buf.String())
}

func setTargetInfoFields(t *testing.T, fields []targetInfoField) {
	t.Helper()

	originalFields := targetInfoFields
	targetInfoFields = fields
	t.Cleanup(func() {
		targetInfoFields = originalFields
	})
}

func newTestField(key string, value string, err error) targetInfoField {
	return targetInfoField{
		key: key,
		collect: func() (string, error) {
			return value, err
		},
	}
}

func expectedLogLine(level string, message string) string {
	return fmt.Sprintf("| %-5s | %-15s | %s", level, testStep, message)
}

func expectedStartLine() string {
	return expectedLogLine("INFO", "installer initial target machine information:")
}

func expectedFieldDebugLine(key string, value string) string {
	return expectedLogLine("DEBUG", fmt.Sprintf("target machine info: %s=%s", key, value))
}

func expectedFieldWarnLine(key string, err error) string {
	return expectedLogLine("WARN", fmt.Sprintf("failed to collect target machine info field(%s): %v", key, err))
}

func splitLogLines(output string) []string {
	trimmedOutput := strings.TrimSpace(output)
	if trimmedOutput == "" {
		return nil
	}

	return strings.Split(trimmedOutput, "\n")
}

func assertEqualStrings(t *testing.T, want []string, got []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("line count = %d, want %d\ngot:\n%s", len(got), len(want), strings.Join(got, "\n"))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d = %q, want %q\nall logs:\n%s", i, got[i], want[i], strings.Join(got, "\n"))
		}
	}
}

func assertNoWarnLog(t *testing.T, logs []string) {
	t.Helper()

	for _, line := range logs {
		if strings.Contains(line, "| WARN ") {
			t.Fatalf("unexpected warn log: %s", line)
		}
	}
}
