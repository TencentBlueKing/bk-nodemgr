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
	"net/url"
	"testing"
)

func TestMaskURLPreservesQueryByDefault(t *testing.T) {
	rawURL := "http://example.com/login/accounts/get_user/?bk_token=bkcrypt%2Bgabcdef%3D&foo=bar&token=short"

	got := maskURL(rawURL, nil)
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	query := parsed.Query()
	if query.Get("bk_token") != "bkcrypt+gabcdef=" {
		t.Fatalf("expected bk_token to be preserved by default, got %q", query.Get("bk_token"))
	}
	if query.Get("token") != "short" {
		t.Fatalf("expected token to be preserved by default, got %q", query.Get("token"))
	}
}

func TestMaskURLMasksConfiguredQuery(t *testing.T) {
	rawURL := "http://example.com/login/accounts/get_user/?bk_token=bkcrypt%2Bgabcdef%3D&foo=bar&token=short"
	urlQueryMasker := map[string]func(string) string{
		"bk_token": defaultHeaderMasker,
		"token":    defaultHeaderMasker,
	}

	got := maskURL(rawURL, urlQueryMasker)
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("failed to parse masked URL: %v", err)
	}

	query := parsed.Query()
	if query.Get("foo") != "bar" {
		t.Fatalf("expected non-configured query to be preserved, got %q", query.Get("foo"))
	}
	if query.Get("bk_token") != "bkc***ef=" {
		t.Fatalf("expected configured bk_token to use default masking, got %q", query.Get("bk_token"))
	}
	if query.Get("token") != "*****" {
		t.Fatalf("expected configured short token to be fully masked, got %q", query.Get("token"))
	}
}
