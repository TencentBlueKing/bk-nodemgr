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

package downloader

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

func normalizeHost(host string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
}

func validateDownloadFilename(filename string) (string, error) {
	if filename == "" {
		return "", errors.New("filename must not be empty")
	}

	if filepath.IsAbs(filename) || strings.ContainsAny(filename, `/\`) || strings.ContainsRune(filename, 0) {
		return "", errors.New("filename must be a base name")
	}

	baseName := filepath.Base(filename)
	if baseName == "." || baseName == ".." || baseName != filename {
		return "", errors.New("filename must be a base name")
	}

	return filename, nil
}

// validateDownloadURL validates rawURL and returns a normalized URL for the request.
// Fragments are removed, the scheme is lowercased, and credentials, bare IPv6,
// IPv6 zones, and invalid ports are rejected.
func validateDownloadURL(rawURL string, allowHosts map[string]struct{}, blockHosts map[string]struct{}) (
	*url.URL, error) {

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, errors.New("URL is invalid")
	}

	parsedURL.Scheme = strings.ToLower(parsedURL.Scheme)
	switch parsedURL.Scheme {
	case httpDownloaderScheme, httpsDownloaderScheme:
	default:
		return nil, fmt.Errorf("URL scheme must be http or https, unsupported scheme: %s", parsedURL.Scheme)
	}

	if parsedURL.User != nil {
		return nil, errors.New("URL credentials are not allowed")
	}

	parsedURL.Fragment = ""
	parsedURL.RawFragment = ""

	if parsedURL.Opaque != "" || !parsedURL.IsAbs() {
		return nil, errors.New("URL must be absolute")
	}

	if strings.Count(parsedURL.Host, ":") > 1 && !strings.HasPrefix(parsedURL.Host, "[") {
		return nil, errors.New("URL IPv6 address must be enclosed in brackets")
	}

	hostname := normalizeHost(parsedURL.Hostname())
	if hostname == "" {
		return nil, errors.New("URL hostname is required")
	}
	if strings.Contains(hostname, "%") {
		return nil, errors.New("URL IPv6 zones are not allowed")
	}

	if err := allowDownloadHost(hostname, allowHosts); err != nil {
		return nil, err
	}

	if err := rejectBlockedHost(hostname, blockHosts); err != nil {
		return nil, err
	}

	return parsedURL, nil
}

func allowDownloadHost(hostname string, allowHosts map[string]struct{}) error {
	if len(allowHosts) == 0 {
		return nil
	}

	if hostname == "" {
		return errors.New("hostname is empty")
	}

	if _, ok := allowHosts[hostname]; !ok {
		return errors.New("URL hostname is not in the download allowlist")
	}

	return nil
}

func rejectBlockedHost(hostname string, blockHosts map[string]struct{}) error {
	if len(blockHosts) == 0 {
		return nil
	}

	if hostname == "" {
		return errors.New("hostname is empty")
	}

	if _, ok := blockHosts[hostname]; ok {
		return errors.New("URL hostname is in the download block list")
	}

	return nil
}

// wrapError wraps cause with errType and a descriptive stage so callers can
// match the sentinel with errors.Is.
func wrapError(errType error, stage string, cause error) error {
	return fmt.Errorf("%w: %s: %w", errType, stage, cause)
}
