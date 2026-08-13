/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package downloader downloads remote HTTP(S) content through the shared rest
// client, buffering the content into a temporary file and verifying its
// checksum before the content is exposed.
package downloader

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tmp"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
)

const (
	httpDownloaderScheme  = "http"
	httpsDownloaderScheme = "https"

	clientNameDownloader = "downloader"

	// DefaultMaxBytes is the default and maximum download size (1 GiB).
	DefaultMaxBytes int64 = 1 << 30
)

// ChecksumAlgorithm identifies a supported digest algorithm for checksum verification.
type ChecksumAlgorithm string

const (
	// ChecksumAlgorithmMD5 verifies content with MD5 (32 hex characters).
	ChecksumAlgorithmMD5 ChecksumAlgorithm = "md5"
)

// Checksum describes an expected digest for downloaded content.
// The checksum verifies artifact integrity compatibility, not authenticity or security.
type Checksum struct {
	// Algorithm selects the digest algorithm; Value must match its hex length.
	// It must be set explicitly: the zero value fails the download with an
	// "unsupported checksum algorithm" error instead of skipping verification.
	Algorithm ChecksumAlgorithm
	// Value is the expected digest in lowercase hexadecimal.
	Value string
}

// DownloadOptions configures one Download call.
type DownloadOptions struct {
	// Filename is the required name used for the temporary file.
	Filename string
	// Checksum is the expected digest of the downloaded content. The content is
	// buffered into a temporary file, verified against this digest, and only
	// then exposed as a fileiface.File.
	Checksum Checksum
}

type client struct {
	allowHosts map[string]struct{}
	blockHosts map[string]struct{}
	maxBytes   int64
	traceSvc   tracing.IService
}

// New builds a downloader for remote package artifacts with the configured host policies.
// Each request builds a rest client for its origin. Redirects are rejected so
// the trust boundary stays on the initially validated URL.
func New(conf config.Downloader) (IDownloader, error) {
	allowHosts := make(map[string]struct{})
	for _, entry := range conf.AllowHosts {
		allowHosts[normalizeHost(entry)] = struct{}{}
	}

	blockHosts := make(map[string]struct{})
	for _, entry := range conf.BlockHosts {
		blockHosts[normalizeHost(entry)] = struct{}{}
	}

	traceSvc, err := tracing.G().NewService(tracing.ServiceConfig{
		ServiceName:     conf.TraceServiceName,
		ServiceCategory: tracing.ServiceCategoryHTTP,
		SampleRate:      conf.TraceSampleRate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to new trace service: %w", err)
	}

	maxBytes := conf.MaxBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}

	return &client{
		allowHosts: allowHosts,
		blockHosts: blockHosts,
		maxBytes:   maxBytes,
		traceSvc:   traceSvc,
	}, nil
}

var (
	// ErrChecksumMismatch is returned when downloaded content does not match the expected checksum.
	ErrChecksumMismatch = errors.New("checksum mismatch")
	// ErrDownloadFailed is returned for any other download failure.
	ErrDownloadFailed = errors.New("download failed")
	// ErrDownloadTooLarge is returned when content exceeds the configured byte limit.
	ErrDownloadTooLarge = errors.New("download too large")
)

var _ IHandler = (*client)(nil)

// Download retrieves rawURL, buffers the content into a temporary file,
// verifies it against opts.Checksum, and returns the verified file. Closing
// the reader returned by Content removes the temporary file.
func (cli *client) Download(nCtx contextx.IContext, rawURL string, opts DownloadOptions) (fileiface.File, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("%w: context must not be nil", ErrDownloadFailed)
	}

	filename, err := validateDownloadFilename(opts.Filename)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDownloadFailed, err)
	}

	parsedURL, err := validateDownloadURL(rawURL, cli.allowHosts, cli.blockHosts)
	if err != nil {
		return nil, wrapError(ErrDownloadFailed, "validate URL", err)
	}

	response, err := cli.fetch(nCtx, parsedURL)
	if err != nil {
		return nil, wrapError(ErrDownloadFailed, "fetch content", err)
	}

	file, err := cli.storeVerifiedContent(nCtx, filename, response, opts.Checksum)
	if err != nil {
		return nil, wrapError(ErrDownloadFailed, "verify content", err)
	}

	return file, nil
}

func (cli *client) fetch(nCtx contextx.IContext, targetURL *url.URL) (*restserver.StreamResponse, error) {
	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{InsecureSkipVerify: true})
	if err != nil {
		return nil, fmt.Errorf("failed to create http client for download host %s: %w", targetURL.Hostname(), err)
	}
	httpClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return fmt.Errorf("%w: redirects are not allowed", ErrDownloadFailed)
	}

	origin := &url.URL{Scheme: targetURL.Scheme, Host: targetURL.Host}
	capability := &restclient.Capability{
		Name:                 clientNameDownloader,
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery(clientNameDownloader, []string{origin.String()}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		TraceSvc:             cli.traceSvc,
	}

	restClient, err := restclient.NewClient(capability, "/")
	if err != nil {
		return nil, fmt.Errorf("failed to create rest client for download host %s: %w", targetURL.Hostname(), err)
	}

	result := restClient.Get().
		SubResourcef(targetURL.Path).
		WithParamsFromURL(targetURL).
		WithContext(nCtx).
		Do()
	reader, err := result.RawStream()
	if err != nil {
		return nil, fmt.Errorf("failed to do get request: %w", err)
	}

	return &restserver.StreamResponse{
		Data:       reader,
		StatusCode: result.StatusCode,
		Headers:    result.Header,
	}, nil
}

func (cli *client) storeVerifiedContent(
	nCtx contextx.IContext,
	filename string,
	response *restserver.StreamResponse,
	expected Checksum,
) (fileiface.File, error) {

	if err := cli.rejectOversizedContentLength(response); err != nil {
		_ = response.Data.Close()

		return nil, err
	}

	var file fileiface.File
	responseBody := newLimitedReadCloser(response.Data, cli.maxBytes+1)
	tmpFile, err := tmp.NewTempFileWithSpecialName(responseBody, filename)
	if err != nil {
		_ = responseBody.Close()

		return nil, fmt.Errorf("failed to create temporary file with special name: %w", err)
	}

	// Remove the temporary file and its directory on every error path; the
	// success path attaches the cleanup to the returned file instead.
	defer func() {
		if file == nil {
			logTempFileCleanupError(tmpFile.Path(), tmpFile.CleanUp())
		}
	}()

	group, err := local.NewLocalDir(filepath.Dir(tmpFile.Path()))
	if err != nil {
		return nil, wrapError(ErrDownloadFailed, "open temporary directory", err)
	}

	file, err = group.GetFile(nCtx, filepath.Base(tmpFile.Path()))
	if err != nil {
		return nil, wrapError(ErrDownloadFailed, "open stored file", err)
	}

	if file.Info().Size > cli.maxBytes {
		return nil, wrapError(ErrDownloadTooLarge, "verify stored file",
			fmt.Errorf("file size(%d) exceeds max bytes(%d)", file.Info().Size, cli.maxBytes))
	}

	switch expected.Algorithm {
	case ChecksumAlgorithmMD5:
		if expected.Value != file.Info().MD5 {
			return nil, wrapError(ErrChecksumMismatch, "verify response body", errors.New("checksum mismatch"))
		}
	default:
		return nil, wrapError(ErrDownloadFailed, "verify stored file", errors.New("unsupported checksum algorithm"))
	}

	return newTemporaryFile(file, tmpFile.Path(), tmpFile.CleanUp), nil
}

func (cli *client) rejectOversizedContentLength(response *restserver.StreamResponse) error {
	// Fail fast when the response advertises a Content-Length larger than the
	// limit; the size check on the stored file remains the authoritative guard
	// because Content-Length can be absent or spoofed.
	contentLength := response.Headers.Get("Content-Length")
	if contentLength == "" {
		return nil
	}

	length, ok := parseContentLength(contentLength)
	if !ok || length <= cli.maxBytes {
		return nil
	}

	return wrapError(ErrDownloadTooLarge, "check content length",
		fmt.Errorf("content length(%d) exceeds max bytes(%d)", length, cli.maxBytes))
}

func parseContentLength(contentLength string) (int64, bool) {
	length, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return 0, false
	}

	return length, true
}
