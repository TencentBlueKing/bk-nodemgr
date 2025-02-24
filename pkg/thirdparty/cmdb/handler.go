package cmdb

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// This document is responsible for processing the conversion of original requests and responses
// from the third-party system and the internal data of the nodeman system.

// Handler the handler of cmdb.
type Handler interface {
	// ListBizHosts list biz hosts
	ListBizHosts(ctx context.Context, BizID int64, page types.Page) ([]*types.Host, error)

	// SearchBusiness search business
	SearchBusiness(ctx context.Context, page types.Page) ([]*types.Business, error)

	// SearchNetworkArea search network area
	SearchNetworkArea(ctx context.Context, page types.Page) ([]*types.NetworkArea, error)
}

type handler struct {
	cli *cli
}

// New initialize a new cmdb handler.
func New(c *client.Capability, conf *Config) (Handler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &handler{cli: cli}, nil
}

// ListBizHosts list biz hosts
func (h *handler) ListBizHosts(ctx context.Context, BizID int64, page types.Page) ([]*types.Host, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListBizHostsReq{
		TenantID: tenantID,
		BKBizID:  BizID,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.listBizHosts(ctx, req)
	if err != nil {
		return nil, err
	}

	hosts := make([]*types.Host, len(resp.Info))
	for idx, host := range resp.Info {
		hosts[idx] = &types.Host{
			TenantID:      tenantID,
			NetworkAreaID: host.BKCloudID,
			BizID:         req.BKBizID,
			HostID:        host.BKHostID,
			InnerIP:       host.BKHostInnerIPV4,
			Mac:           host.BKMac,
			OSType:        host.BKOsType,
		}
	}

	return hosts, nil
}

// SearchBusiness search business
func (h *handler) SearchBusiness(ctx context.Context, page types.Page) ([]*types.Business, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &SearchBusinessReq{
		TenantID:          tenantID,
		BKSupplierAccount: "tencent",
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
		Fields: nil,
	}

	resp, err := h.cli.searchBusiness(ctx, req)
	if err != nil {
		return nil, err
	}

	bizs := make([]*types.Business, len(resp.Info))
	for idx, business := range resp.Info {
		bizs[idx] = &types.Business{
			TenantID: tenantID,
			BizID:    business.BKBizID,
			BizName:  business.BKBizName,
		}
	}

	return bizs, nil
}

// SearchNetworkArea search network area
func (h *handler) SearchNetworkArea(ctx context.Context, page types.Page) ([]*types.NetworkArea, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &SearchCloudAreaReq{
		TenantID: tenantID,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.searchCloudArea(ctx, req)
	if err != nil {
		return nil, err
	}

	netAreas := make([]*types.NetworkArea, len(resp.Info))
	for idx, netArea := range resp.Info {
		netAreas[idx] = &types.NetworkArea{
			TenantID: tenantID,
			ID:       netArea.BkCloudID,
			Name:     netArea.BkCloudName,
		}
	}

	return netAreas, nil
}
