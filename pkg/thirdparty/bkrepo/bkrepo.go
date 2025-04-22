package bkrepo

import (
	"context"
	"fmt"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	"net/http"
	"net/url"
	"strconv"
)

// This file only supports requesting and getting responses.

// HeaderSetter get auth header.
type HeaderSetter interface {
	GetAuthHeader() (string, error)
}

// Config the config of gse.
type Config struct {
	HeaderSetter HeaderSetter
	Endpoint     string
	RepoName     string
	ProjectID    string
}

// cli client for gse.
type cli struct {
	client rest.ClientInterface
	config *Config
}

// newClient initialize a new gse client.
func newClient(c *client.Capability, conf *Config) (*cli, error) {
	restCli, err := rest.NewClient(c, "/")
	if err != nil {
		return nil, err
	}

	return &cli{
		client: restCli,
		config: conf,
	}, nil
}

// getCommonHeader get gse common header.
func (c *cli) getCommonHeader() (http.Header, error) {
	header := http.Header{}

	header.Set(restheader.BKRIDKey, restheader.BKRIDGenerator())
	authHeader, err := c.config.HeaderSetter.GetAuthHeader()
	if err != nil {
		return nil, err
	}

	// notice: bkrepo didn't use restheader.BKGWAuthKey, it uses restheader.AuthKey.
	header.Set(restheader.AuthKey, authHeader)

	return header, nil
}

// DownloadFile download file from bkrepo.
func (c *cli) DownloadFile(ctx context.Context, req *DownloadFileReq) (*DownloadFileResp, error) {
	resp := new(DownloadFileResp)
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	subPath, err := url.JoinPath("generic", req.ProjectID, req.RepoName, req.Path)
	if err != nil {
		return nil, err
	}

	resp.Data, err = c.client.Get().
		SubResourcef(subPath).
		WithParam("download", "true").
		WithContext(ctx).
		WithHeaders(header).
		Do().RawData()
	if err != nil {
		return nil, fmt.Errorf("download file failed, err: %w", err)
	}

	return resp, nil
}

// UploadFile upload file to bkrepo.
func (c *cli) UploadFile(ctx context.Context, req *UploadFileReq) (*UploadFileResp, error) {
	resp := new(BaseBroker[*UploadFileResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, fmt.Errorf("upload file failed, err: %w", err)
	}
	header = req.Info.BindHeader(header)

	subPath, err := url.JoinPath("generic", c.config.ProjectID, c.config.RepoName, req.Path)
	if err != nil {
		return nil, fmt.Errorf("upload file failed, err: %w", err)
	}

	err = c.client.Put().
		SubResourcef(subPath).
		WithContext(ctx).
		WithHeaders(header).
		RawBody(req.Data).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("upload file failed, err: %w", err)
	}

	return resp.Data, nil
}

// QueryNodeInfo query node info.
func (c *cli) QueryNodeInfo(ctx context.Context, req *QueryNodeInfoReq) (*QueryNodeInfoResp, error) {
	resp := new(BaseBroker[*QueryNodeInfoResp])

	header, err := c.getCommonHeader()
	if err != nil {
		return nil, fmt.Errorf("query node info failed, err: %w", err)
	}

	subPath, err := url.JoinPath("/repository/api/node/detail", c.config.ProjectID, c.config.RepoName, req.Path)
	if err != nil {
		return nil, fmt.Errorf("query node info failed, err: %w", err)
	}

	err = c.client.Put().
		SubResourcef(subPath).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("query node info failed, err: %w", err)
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("query node info failed, err: %w", err)
	}

	return resp.Data, nil
}

// ListNode list nodes request.
func (c *cli) ListNode(ctx context.Context, req *ListNodeReq) (*ListNodeResp, error) {
	resp := new(BaseBroker[*ListNodeResp])

	header, err := c.getCommonHeader()
	if err != nil {
		return nil, fmt.Errorf("list node failed, err: %w", err)
	}

	subPath, err := url.JoinPath("/repository/api/node/page", c.config.ProjectID, c.config.RepoName, req.Path)
	if err != nil {
		return nil, fmt.Errorf("list node failed, err: %w", err)
	}

	err = c.client.Put().
		SubResourcef(subPath).
		WithContext(ctx).
		WithHeaders(header).
		WithParams(map[string]string{
			"pageSize":        strconv.Itoa(req.PageSize),
			"pageNum":         strconv.Itoa(req.PageNum),
			"includeFolder":   "true",
			"includeMetadata": "false",
			"sortProperty":    "id",
			"direction":       "DESC",
		}).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("list node failed, err: %w", err)
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list node failed, err: %w", err)
	}

	return resp.Data, nil
}
