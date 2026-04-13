/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package versionlog

import "testing"

func TestNormalizeLanguage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "empty defaults to zh", input: "", expected: defaultLanguage},
		{name: "zh locale maps to zh", input: "zh-cn", expected: defaultLanguage},
		{name: "en locale maps to en", input: "en-us", expected: englishLanguage},
		{name: "unknown defaults to zh", input: "ja-jp", expected: defaultLanguage},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actual := normalizeLanguage(testCase.input)
			if actual != testCase.expected {
				t.Fatalf("normalizeLanguage(%q) = %q, want %q", testCase.input, actual, testCase.expected)
			}
		})
	}
}

func TestBuildVersionLogEntry(t *testing.T) {
	t.Parallel()

	entry, ok := buildVersionLogEntry("v3.0.1-alpha.17_2026-04-10.md")
	if !ok {
		t.Fatal("expected markdown changelog entry to be parsed")
	}

	if entry.Version != "v3.0.1-alpha.17" {
		t.Fatalf("unexpected version: %s", entry.Version)
	}

	if entry.Date != "2026-04-10" {
		t.Fatalf("unexpected date: %s", entry.Date)
	}

	if entry.IsCurrent {
		t.Fatal("entry should not be marked as current in this test")
	}
}

func TestCompareVersions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		left     string
		right    string
		expected int
	}{
		{name: "newer stable first", left: "v3.0.1", right: "v3.0.0", expected: -1},
		{name: "prerelease before stable", left: "v3.0.1-rc.1", right: "v3.0.1", expected: -1},
		{name: "higher prerelease rank first", left: "v3.0.1-beta.1", right: "v3.0.1-alpha.1", expected: -1},
		{name: "lower mark first", left: "v3.0.1-1", right: "v3.0.1-2", expected: -1},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actual := compareVersions(testCase.left, testCase.right)
			if actual != testCase.expected {
				t.Fatalf("compareVersions(%q, %q) = %d, want %d", testCase.left, testCase.right, actual, testCase.expected)
			}
		})
	}
}
