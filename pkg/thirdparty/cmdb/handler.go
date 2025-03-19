package cmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// This document is responsible for processing the conversion of original requests and responses
// from the third-party system and the internal data of the nodeman system.

// Handler the handler of cmdb.
type Handler interface {
	// ListBizHosts list biz hosts
	ListBizHosts(ctx context.Context, bizID int64, page types.Page) ([]*types.Host, error)

	// ListBizHostsTopo list biz hosts topo
	ListBizHostsTopo(ctx context.Context, bizID int64, page types.Page) ([]*types.HostWithHostTopo, error)

	// ListHostsWithoutBiz list hosts without biz
	ListHostsWithoutBiz(ctx context.Context, page types.Page) ([]*types.Host, error)

	// FindHostRelationsByBiz find host relations by biz
	FindHostRelationsByBiz(ctx context.Context, bizID int64, hostIDs []int64) ([]*types.HostTopoRelation, error)

	// BindHostAgent bind host agent
	BindHostAgent(ctx context.Context, hostAgentID []*types.HostAgentID) error

	// UnbindHostAgent bind host agent
	UnbindHostAgent(ctx context.Context, hostAgentID []*types.HostAgentID) error

	// AddHostToBusinessIdle add host to business idle
	AddHostToBusinessIdle(ctx context.Context, bizID int64, hosts []*types.CreateHostInfo) ([]int64, error)

	// PushHostIdentifier push host identifier
	PushHostIdentifier(ctx context.Context, hostIDs []int64) (*types.PushHostIdentifierTaskInfo, error)

	// FindHostIdentifierPushResult find host identifier push result
	FindHostIdentifierPushResult(ctx context.Context, taskID string) (*types.PushHostIdentifiersTaskResult, error)

	// SearchBusiness search business
	SearchBusiness(ctx context.Context, page types.Page) ([]*types.Business, error)

	// SearchNetworkArea search network area
	SearchNetworkArea(ctx context.Context, page types.Page) ([]*types.NetworkArea, error)

	// CreateNetworkArea create network area
	CreateNetworkArea(ctx context.Context, networkAreaName string, cloudVendor string) (*types.NetworkArea, error)

	// UpdateNetworkArea update network area
	UpdateNetworkArea(ctx context.Context, id int64, networkAreaName string, cloudVendor string) error

	// DeleteNetworkArea delete network area
	DeleteNetworkArea(ctx context.Context, id int64) error

	// UpdateHostNetworkAreaField update host network area field
	UpdateHostNetworkAreaField(ctx context.Context, hostIDs []int64, bizID int64, networkAreaID int64) error

	// SearchCloudVendor search cloud vendor
	SearchCloudVendor(ctx context.Context) ([]*types.CloudVendor, error)

	// SearchOsType search os type
	SearchOsType(ctx context.Context) ([]*types.OsType, error)

	// ExecuteHostDynamicGroup execute host dynamic group
	ExecuteHostDynamicGroup(ctx context.Context, bizID int64, groupID string, page types.Page) ([]*types.Host, error)

	// SearchDynamicGroup search dynamic group
	SearchDynamicGroup(ctx context.Context, bizID int64, page types.Page) ([]*types.DynamicGroup, error)

	// ListServiceTemplate list service template
	ListServiceTemplate(ctx context.Context, bizID int64, page types.Page) ([]*types.ServiceTemplate, error)

	// FindHostByServiceTemplate find host by service template
	FindHostByServiceTemplate(ctx context.Context, bizID int64, serviceTemplateIDs []int64, page types.Page) (
		[]*types.Host, error)

	// ListServiceInstanceDetail list service instance detail
	ListServiceInstanceDetail(ctx context.Context, bizID int64, page types.Page) ([]*types.ServiceInstanceDetail, error)

	// ListServiceInstanceByHost list service instance by host
	ListServiceInstanceByHost(ctx context.Context, bizID int64, hostID int64, page types.Page) (
		[]*types.ServiceInstance, error)

	// ListProcessInstance list process instance
	ListProcessInstance(ctx context.Context, bizID int64, serviceInstanceID int64) ([]*types.ProcessInstance, error)

	// ListSetTemplate list set template
	ListSetTemplate(ctx context.Context, bizID int64, page types.Page) ([]*types.SetTemplate, error)

	// ListServiceInstanceBySetTemplate list service instance by set template
	ListServiceInstanceBySetTemplate(ctx context.Context, bizID int64, setTemplate int64, page types.Page) (
		[]*types.ServiceInstance, error)

	// ListProcTemplate list process instance by set template
	ListProcTemplate(ctx context.Context, bizID int64, serviceTemplateID int64, processTemplateIDs []int64) (
		[]*types.ProcessTemplate, error)

	// FindHostBySetTemplate find host by set template
	FindHostBySetTemplate(ctx context.Context, bizID int64, setTemplateIDs []int64, page types.Page) ([]*types.Host,
		error)
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

// ListBizHosts list biz hosts.
func (h *handler) ListBizHosts(ctx context.Context, bizID int64, page types.Page) ([]*types.Host, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListBizHostsReq{
		TenantID: tenantID,
		BKBizID:  bizID,
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
			HostID:   host.BKHostID,
			TenantID: tenantID,
			Static: &types.HostStatic{
				BizID:         bizID,
				NetworkAreaID: host.BKCloudID,
				HostName:      host.BKHostName,
				DeptName:      host.DeptName,
				InnerIP:       host.BKHostInnerIPV4,
				InnerIPV6:     host.BKHostInnerIPV6,
				OuterIP:       host.BKHostOuterIPV4,
				OuterIPV6:     host.BKHostOuterIPV6,
				Mac:           host.BKMac,
				OSType:        host.BKOsType,
				SyncedAgentID: host.BKAgentID,
			},
			Dynamic: types.NewBlankNodeDynamic(),
		}
	}

	return hosts, nil
}

// ListBizHostsTopo list biz hosts topo.
func (h *handler) ListBizHostsTopo(ctx context.Context, bizID int64, page types.Page) (
	[]*types.HostWithHostTopo, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListBizHostsTopoReq{
		TenantID: tenantID,
		BKBizID:  bizID,
		Fields:   DefaultSearchHostFields,
		Page: Page{
			Start: page.Offset,
			Sort:  page.Sort,
			Limit: page.Limit,
		},
	}
	resp, err := h.cli.listBizHostsTopo(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.HostWithHostTopo, len(resp.Info))
	for index, host := range resp.Info {
		result[index] = &types.HostWithHostTopo{
			Host: &types.Host{
				HostID:   host.Host.BKHostID,
				TenantID: tenantID,
				Static: &types.HostStatic{
					BizID:         bizID,
					NetworkAreaID: host.Host.BKCloudID,
					HostName:      host.Host.BKHostName,
					DeptName:      host.Host.DeptName,
					InnerIP:       host.Host.BKHostInnerIPV4,
					InnerIPV6:     host.Host.BKHostInnerIPV6,
					OuterIP:       host.Host.BKHostOuterIPV4,
					OuterIPV6:     host.Host.BKHostOuterIPV6,
					Mac:           host.Host.BKMac,
					OSType:        host.Host.BKOsType,
					SyncedAgentID: host.Host.BKAgentID,
				},
				Dynamic: types.NewBlankNodeDynamic(),
			},
		}
		result[index].HostTopo = make([]*types.HostTopo, len(host.Topo))
		for topoIndex, topo := range host.Topo {
			result[index].HostTopo[topoIndex] = &types.HostTopo{
				SetID:   topo.BKSetID,
				SetName: topo.BKSetName,
			}
			result[index].HostTopo[topoIndex].Module = make([]*struct {
				ModuleID   int64
				ModuleName string
			}, len(topo.Module))
			for moduleIndex, module := range topo.Module {
				result[index].HostTopo[topoIndex].Module[moduleIndex] = &struct {
					ModuleID   int64
					ModuleName string
				}{
					ModuleID:   module.BKModuleID,
					ModuleName: module.BKModuleName,
				}
			}
		}
	}

	return result, nil
}

func (h *handler) ListHostsWithoutBiz(ctx context.Context, page types.Page) ([]*types.Host, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListHostsWithoutBusinessReq{
		TenantID: tenantID,
		Fields:   DefaultSearchHostFields,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.listHostsWithoutBusiness(ctx, req)
	if err != nil {
		return nil, err
	}

	hosts := make([]*types.Host, len(resp.Info))
	for idx, host := range resp.Info {
		hosts[idx] = &types.Host{
			HostID:   host.BKHostID,
			TenantID: tenantID,
			Static: &types.HostStatic{
				NetworkAreaID: host.BKCloudID,
				HostName:      host.BKHostName,
				DeptName:      host.DeptName,
				InnerIP:       host.BKHostInnerIPV4,
				InnerIPV6:     host.BKHostInnerIPV6,
				OuterIP:       host.BKHostOuterIPV4,
				OuterIPV6:     host.BKHostOuterIPV6,
				Mac:           host.BKMac,
				OSType:        host.BKOsType,
				SyncedAgentID: host.BKAgentID,
			},
			Dynamic: types.NewBlankNodeDynamic(),
		}
	}

	return hosts, nil
}

// FindHostRelationsByBiz find host relations by biz.
func (h *handler) FindHostRelationsByBiz(ctx context.Context, bizID int64, hostIDs []int64) ([]*types.HostTopoRelation,
	error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &FindHostBizRelationsReq{
		TenantID: tenantID,

		BKBizID:  bizID,
		BKHostID: hostIDs,
	}

	resp, err := h.cli.findHostBizRelations(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.HostTopoRelation, len(*resp))
	for index, relation := range *resp {
		result[index] = &types.HostTopoRelation{
			BizID:    relation.BKBizID,
			HostID:   relation.BKHostID,
			SetID:    relation.BKSetID,
			ModuleID: relation.BKModuleID,
		}
	}

	return result, nil
}

// BindHostAgent bind host agent.
func (h *handler) BindHostAgent(ctx context.Context, hostAgentID []*types.HostAgentID) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &BindHostAgentReq{TenantID: tenantID}
	for _, hostAgent := range hostAgentID {
		req.List = append(req.List, &HostAgentIDInfo{
			BKHostID:  hostAgent.HostID,
			BKAgentID: hostAgent.AgentID,
		})
	}
	err = h.cli.bindHostAgent(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// UnbindHostAgent bind host agent.
func (h *handler) UnbindHostAgent(ctx context.Context, hostAgentID []*types.HostAgentID) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &UnbindHostAgentReq{TenantID: tenantID}
	for _, hostAgent := range hostAgentID {
		req.List = append(req.List, &HostAgentIDInfo{
			BKHostID:  hostAgent.HostID,
			BKAgentID: hostAgent.AgentID,
		})
	}
	err = h.cli.unbindHostAgent(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// AddHostToBusinessIdle add host to business idle.
func (h *handler) AddHostToBusinessIdle(ctx context.Context, bizID int64, hosts []*types.CreateHostInfo) (
	[]int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &AddHostToBusinessIdleReq{TenantID: tenantID, BKBizID: bizID}
	for _, host := range hosts {
		req.BKHostList = append(req.BKHostList, &CreateHostInfo{
			BKCloudID:         host.NetworkAreaID,
			BKHostInnerIP:     host.InnerIP,
			BKHostInnerIPV6:   host.InnerIPV6,
			BKHostOuterIP:     host.OuterIP,
			BKHostOuterIPV6:   host.OuterIPV6,
			BKOSType:          host.OSType,
			BKCpuArchitecture: host.Arch,
			BKAddressing:      string(host.Addressing),
		})
	}
	resp, err := h.cli.addHostToBusinessIdle(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.BKHostIDs, nil
}

// PushHostIdentifier push host identifier.
func (h *handler) PushHostIdentifier(ctx context.Context, hostIDs []int64) (
	*types.PushHostIdentifierTaskInfo, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &PushHostIdentifierReq{
		TenantID:  tenantID,
		BKHostIDs: hostIDs,
	}
	resp, err := h.cli.pushHostIdentifier(ctx, req)
	if err != nil {
		return nil, err
	}

	result := &types.PushHostIdentifierTaskInfo{
		TaskID: resp.TaskID,
	}
	result.HostInfo = make([]*types.HostIdentification, len(resp.HostInfos))
	for index, hostInfo := range resp.HostInfos {
		result.HostInfo[index] = &types.HostIdentification{
			HostID:         hostInfo.BKHostID,
			Identification: hostInfo.Identification,
		}
	}

	return result, nil
}

// FindHostIdentifierPushResult find host identifier push result.
func (h *handler) FindHostIdentifierPushResult(ctx context.Context, taskID string) (
	*types.PushHostIdentifiersTaskResult, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &FindHostIdentifierPushResultReq{
		TenantID: tenantID,

		TaskID: taskID,
	}
	resp, err := h.cli.findHostIdentifierPushResult(ctx, req)
	if err != nil {
		return nil, err
	}

	result := &types.PushHostIdentifiersTaskResult{
		SuccessHostIDList: resp.SuccessList,
		FailedHostIDList:  resp.FailedList,
		PendingHostIDList: resp.PendingList,
	}

	return result, nil
}

// SearchBusiness search business.
func (h *handler) SearchBusiness(ctx context.Context, page types.Page) ([]*types.Business, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &SearchBusinessReq{
		TenantID: tenantID,
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

// SearchNetworkArea search network area.
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
			TenantID:    tenantID,
			ID:          netArea.BKCloudID,
			Name:        netArea.BKCloudName,
			CloudVendor: netArea.BKCloudVendor,
		}
	}

	return netAreas, nil
}

// CreateNetworkArea create network area.
func (h *handler) CreateNetworkArea(ctx context.Context, networkAreaName string, cloudVendor string) (
	*types.NetworkArea, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &CreateCloudAreaReq{
		TenantID:      tenantID,
		BKCloudName:   networkAreaName,
		BKCloudVendor: cloudVendor,
	}

	resp, err := h.cli.createCloudArea(ctx, req)
	if err != nil {
		return nil, err
	}

	netArea := &types.NetworkArea{
		TenantID:    tenantID,
		ID:          resp.Created.ID,
		Name:        networkAreaName,
		CloudVendor: cloudVendor,
	}

	return netArea, nil
}

// UpdateNetworkArea update network area.
func (h *handler) UpdateNetworkArea(ctx context.Context, id int64, networkAreaName string, cloudVendor string) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &UpdateCloudAreaReq{
		TenantID:      tenantID,
		BKCloudID:     id,
		BKCloudName:   networkAreaName,
		BKCloudVendor: cloudVendor,
	}

	err = h.cli.updateCloudArea(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// DeleteNetworkArea delete network area.
func (h *handler) DeleteNetworkArea(ctx context.Context, id int64) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &DeleteCloudAreaReq{
		TenantID:  tenantID,
		BKCloudID: id,
	}

	err = h.cli.deleteCloudArea(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// UpdateHostNetworkAreaField update host network area field.
func (h *handler) UpdateHostNetworkAreaField(ctx context.Context, hostIDs []int64, bizID int64,
	networkAreaID int64) error {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &UpdateHostCloudAreaFieldReq{
		TenantID:  tenantID,
		BKBizID:   bizID,
		BKCloudID: networkAreaID,
		BKHostIDs: hostIDs,
	}

	err = h.cli.updateHostCloudAreaField(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// searchObjectAttributeEnumOption search cmdb object attribute's option, like bk_cloud_vendor and bk_os_type.
func (h *handler) searchObjectAttributeEnumOption(ctx context.Context, objID string, bizID int64, objAttrID string) (
	[]*EnumOption, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &SearchObjectAttributeReq{
		TenantID: tenantID,
		BKObjID:  objID,
		BKBizID:  bizID,
	}

	resp, err := h.cli.searchObjectAttribute(ctx, req)
	if err != nil {
		return nil, err
	}

	var result []*EnumOption
	for _, objAttribute := range *resp {
		if objAttribute.BKPropertyID != objAttrID {
			continue
		}

		options, ok := objAttribute.Option.([]any)
		if !ok {
			return nil, fmt.Errorf("try to convert type to []any failed, bk_property_id(%s), option(%v)",
				objAttrID, objAttribute.Option)
		}

		result = make([]*EnumOption, len(options))
		for index, option := range options {
			mapOption, ok := option.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("try to convert type to map[string]any failed, bk_property_id(%s), "+
					"option(%v)", objAttrID, option)
			}

			key, ok := mapOption["id"].(string)
			if !ok {
				return nil, fmt.Errorf("try to convert type to string failed, bk_property_id(%s), option[id](%v)",
					objAttrID, mapOption["id"])
			}

			name, ok := mapOption["name"].(string)
			if !ok {
				return nil, fmt.Errorf("try to convert type to string failed, bk_property_id(%s), option[name](%v)",
					objAttrID, mapOption["name"])
			}

			result[index] = &EnumOption{
				Key:   key,
				Value: name,
			}
		}
	}

	return result, nil
}

// SearchCloudVendor search cloud vendor.
func (h *handler) SearchCloudVendor(ctx context.Context) ([]*types.CloudVendor, error) {
	result, err := h.searchObjectAttributeEnumOption(ctx, "plat", DefaultBusinessID, "bk_cloud_vendor")
	if err != nil {
		return nil, err
	}

	cloudVendors := make([]*types.CloudVendor, len(result))
	for idx, vendor := range result {
		cloudVendors[idx] = &types.CloudVendor{
			Key:  vendor.Key,
			Name: vendor.Value,
		}
	}

	return cloudVendors, nil
}

// SearchOsType search ostype.
func (h *handler) SearchOsType(ctx context.Context) ([]*types.OsType, error) {
	result, err := h.searchObjectAttributeEnumOption(ctx, "host", DefaultBusinessID, "bk_os_type")
	if err != nil {
		return nil, err
	}

	osTypes := make([]*types.OsType, len(result))
	for idx, osType := range result {
		osTypes[idx] = &types.OsType{
			Key:  osType.Key,
			Name: osType.Value,
		}
	}

	return osTypes, nil
}

// ExecuteHostDynamicGroup execute host dynamic group.
func (h *handler) ExecuteHostDynamicGroup(ctx context.Context, bizID int64, groupID string,
	page types.Page) ([]*types.Host, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ExecuteDynamicGroupReq{
		TenantID:       tenantID,
		BKBizID:        bizID,
		ID:             groupID,
		Fields:         DefaultSearchHostFields,
		DisableCounter: false,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	var resp *ExecuteDynamicGroupResp
	resp, err = h.cli.executeDynamicGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Host, len(resp.Info))
	for index, host := range resp.Info {
		jsonData, err := json.Marshal(host)
		if err != nil {
			return nil, err
		}

		var hostData HostInfo
		err = json.Unmarshal(jsonData, &hostData)
		if err != nil {
			return nil, err
		}

		result[index] = &types.Host{
			HostID:   hostData.BKHostID,
			TenantID: tenantID,
			Static: &types.HostStatic{
				BizID:         bizID,
				NetworkAreaID: hostData.BKCloudID,
				HostName:      hostData.BKHostName,
				DeptName:      hostData.DeptName,
				InnerIP:       hostData.BKHostInnerIPV4,
				InnerIPV6:     hostData.BKHostInnerIPV6,
				OuterIP:       hostData.BKHostOuterIPV4,
				OuterIPV6:     hostData.BKHostOuterIPV6,
				Mac:           hostData.BKMac,
				OSType:        hostData.BKOsType,
			},
			Dynamic: types.NewBlankNodeDynamic(),
		}
	}

	return result, nil
}

// SearchDynamicGroup search dynamic group.
func (h *handler) SearchDynamicGroup(ctx context.Context, bizID int64, page types.Page) (
	[]*types.DynamicGroup, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &SearchDynamicGroupReq{
		TenantID: tenantID,
		BKBizID:  bizID,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.searchDynamicGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.DynamicGroup, len(resp.Info))
	for index, group := range resp.Info {
		result[index] = &types.DynamicGroup{
			ID:    group.ID,
			BizID: group.BKBizID,
			ObjID: group.BKObjID,
			Name:  group.Name,
		}
	}

	return result, nil
}

// ListServiceTemplate list service template.
func (h *handler) ListServiceTemplate(ctx context.Context, bizID int64, page types.Page) ([]*types.ServiceTemplate,
	error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListServiceTemplateReq{
		TenantID: tenantID,

		BKBizID: bizID,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.listServiceTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	serviceTemplates := make([]*types.ServiceTemplate, len(resp.Info))
	for index, serviceTemplate := range resp.Info {
		serviceTemplates[index] = &types.ServiceTemplate{
			ID:                  serviceTemplate.ID,
			BizID:               serviceTemplate.BKBizID,
			ServiceTemplateName: serviceTemplate.ServiceTemplateName,
			ServiceCategoryID:   serviceTemplate.ServiceCategoryID,
			HostApplyEnabled:    serviceTemplate.HostApplyEnabled,
		}
	}

	return serviceTemplates, nil
}

// FindHostByServiceTemplate find host by service template.
func (h *handler) FindHostByServiceTemplate(ctx context.Context, bizID int64, serviceTemplateIDs []int64,
	page types.Page) ([]*types.Host, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &FindHostByServiceTemplateReq{
		TenantID: tenantID,

		BKBizID:              bizID,
		BKServiceTemplateIDs: serviceTemplateIDs,
		Fields:               DefaultSearchHostFields,
		Page: Page{
			Start: page.Offset,
			Sort:  page.Sort,
			Limit: page.Limit,
		},
	}

	resp, err := h.cli.findHostByServiceTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Host, len(resp.Info))
	for index, host := range resp.Info {
		result[index] = &types.Host{
			HostID:   host.BKHostID,
			TenantID: tenantID,
			Static: &types.HostStatic{
				BizID:         bizID,
				NetworkAreaID: host.BKCloudID,
				HostName:      host.BKHostName,
				DeptName:      host.DeptName,
				InnerIP:       host.BKHostInnerIPV4,
				InnerIPV6:     host.BKHostInnerIPV6,
				OuterIP:       host.BKHostOuterIPV4,
				OuterIPV6:     host.BKHostOuterIPV6,
				Mac:           host.BKMac,
				OSType:        host.BKOsType,
				SyncedAgentID: host.BKAgentID,
			},
			Dynamic: types.NewBlankNodeDynamic(),
		}
	}

	return result, nil
}

// ListServiceInstanceDetail list service instance detail.
func (h *handler) ListServiceInstanceDetail(ctx context.Context, bizID int64, page types.Page) (
	[]*types.ServiceInstanceDetail, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListServiceInstanceDetailReq{
		TenantID: tenantID,

		BKBizID: bizID,
		Page: Page{
			Start: page.Offset,
			Sort:  page.Sort,
			Limit: page.Limit,
		},
	}
	resp, err := h.cli.listServiceInstanceDetail(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.ServiceInstanceDetail, len(resp.Info))
	for index, serviceInstance := range resp.Info {
		result[index] = &types.ServiceInstanceDetail{
			ID:                  serviceInstance.ID,
			BizID:               serviceInstance.BKBizID,
			ServiceTemplateName: serviceInstance.Name,
			ServiceTemplateID:   serviceInstance.ServiceTemplateID,
			HostID:              serviceInstance.BKHostID,
			ModuleID:            serviceInstance.BKModuleID,
			ServiceCategoryID:   serviceInstance.ServiceCategoryID,
		}

		result[index].ProcessInstances = make([]*types.ProcessInstance, len(serviceInstance.ProcessInstances))
		for svcInstIndex, svcInst := range serviceInstance.ProcessInstances {
			result[index].ProcessInstances[svcInstIndex] = &types.ProcessInstance{
				Property: &types.ProcessProperty{
					AutoStart:       svcInst.Process.AutoStart,
					BizID:           svcInst.Process.BKBizID,
					FuncName:        svcInst.Process.BKFuncName,
					ProcessID:       svcInst.Process.BKProcessID,
					ProcessName:     svcInst.Process.BKProcessName,
					StartParamRegex: svcInst.Process.BKStartParamRegex,
					Description:     svcInst.Process.Description,
					FaceStopCMD:     svcInst.Process.FaceStopCMD,
					PidFile:         svcInst.Process.PidFile,
					Priority:        svcInst.Process.Priority,
					ProcNum:         svcInst.Process.ProcNum,
					ReloadCMD:       svcInst.Process.ReloadCMD,
					RestartCMD:      svcInst.Process.RestartCMD,
					StartCMD:        svcInst.Process.StartCMD,
					StopCMD:         svcInst.Process.StopCMD,
					Timeout:         svcInst.Process.Timeout,
					User:            svcInst.Process.User,
					WorkPath:        svcInst.Process.WorkPath,
				},
				Relation: &types.ProcessRelation{
					BizID:             svcInst.Relation.BKBizID,
					ProcessID:         svcInst.Relation.BKProcessID,
					ServiceInstanceID: svcInst.Relation.ServiceInstanceID,
					ProcessTemplateID: svcInst.Relation.ProcessTemplateID,
					HostID:            svcInst.Relation.BKHostID,
				},
			}
			result[index].ProcessInstances[svcInstIndex].Property.BindInfo = make([]*types.BindInfo,
				len(svcInst.Process.BindInfo))
			for bindInfoIndex, bindInfo := range svcInst.Process.BindInfo {
				result[index].ProcessInstances[svcInstIndex].Property.BindInfo[bindInfoIndex] = &types.BindInfo{
					Enable:        bindInfo.Enable,
					IP:            bindInfo.IP,
					Port:          bindInfo.Port,
					Protocol:      bindInfo.Protocol,
					TemplateRowID: bindInfo.TemplateRowID,
				}
			}
		}
	}

	return result, nil
}

// ListServiceInstanceByHost list service instance by host.
func (h *handler) ListServiceInstanceByHost(ctx context.Context, bizID int64, hostID int64, page types.Page) (
	[]*types.ServiceInstance, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListServiceInstanceByHostReq{
		TenantID: tenantID,

		BKBizID:  bizID,
		BKHostID: hostID,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}
	resp, err := h.cli.listServiceInstanceByHost(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.ServiceInstance, len(resp.Info))
	for index, serviceInstance := range resp.Info {
		result[index] = &types.ServiceInstance{
			ID:                  serviceInstance.ID,
			BizID:               serviceInstance.BKBizID,
			ServiceInstanceName: serviceInstance.ServiceInstanceName,
			HostID:              serviceInstance.BKHostID,
			ModuleID:            serviceInstance.BKModuleID,
		}
	}

	return result, nil
}

// ListProcessInstance list process instance.
func (h *handler) ListProcessInstance(ctx context.Context, bizID int64, serviceInstanceID int64) (
	[]*types.ProcessInstance, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListProcessInstanceReq{
		TenantID:          tenantID,
		BKBizID:           bizID,
		ServiceInstanceID: serviceInstanceID,
	}

	resp, err := h.cli.listProcessInstance(ctx, req)
	if err != nil {
		return nil, err
	}

	processInstances := make([]*types.ProcessInstance, len(*resp))
	for index, processInstance := range *resp {
		processInstances[index] = &types.ProcessInstance{
			Property: &types.ProcessProperty{
				AutoStart:       processInstance.Property.AutoStart,
				BizID:           processInstance.Property.BKBizID,
				FuncName:        processInstance.Property.BKFuncName,
				ProcessID:       processInstance.Property.BKProcessID,
				ProcessName:     processInstance.Property.BKProcessName,
				StartParamRegex: processInstance.Property.BKStartParamRegex,
				Description:     processInstance.Property.Description,
				FaceStopCMD:     processInstance.Property.FaceStopCMD,
				PidFile:         processInstance.Property.PidFile,
				Priority:        processInstance.Property.Priority,
				ProcNum:         processInstance.Property.ProcNum,
				ReloadCMD:       processInstance.Property.ReloadCMD,
				RestartCMD:      processInstance.Property.RestartCMD,
				StartCMD:        processInstance.Property.StartCMD,
				StopCMD:         processInstance.Property.StopCMD,
				Timeout:         processInstance.Property.Timeout,
				User:            processInstance.Property.User,
				WorkPath:        processInstance.Property.WorkPath,
			},
			Relation: &types.ProcessRelation{
				BizID:             processInstance.Relation.BKBizID,
				ProcessID:         processInstance.Relation.BKProcessID,
				ServiceInstanceID: processInstance.Relation.ServiceInstanceID,
				ProcessTemplateID: processInstance.Relation.ProcessTemplateID,
				HostID:            processInstance.Relation.BKHostID,
			},
		}

		processInstances[index].Property.BindInfo = make([]*types.BindInfo, len(processInstance.Property.BindInfo))
		for bindInfoIdx, bindInfo := range processInstance.Property.BindInfo {
			processInstances[index].Property.BindInfo[bindInfoIdx] = &types.BindInfo{
				Enable:        bindInfo.Enable,
				IP:            bindInfo.IP,
				Port:          bindInfo.Port,
				Protocol:      bindInfo.Protocol,
				TemplateRowID: bindInfo.TemplateRowID,
			}
		}
	}

	return processInstances, nil
}

// ListSetTemplate list set template.
func (h *handler) ListSetTemplate(ctx context.Context, bizID int64, page types.Page) (
	[]*types.SetTemplate, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListSetTemplateReq{
		TenantID: tenantID,

		BKBizID: bizID,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}
	resp, err := h.cli.listSetTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.SetTemplate, len(resp.Info))
	for index, setTemplate := range resp.Info {
		result[index] = &types.SetTemplate{
			ID:              setTemplate.ID,
			BizID:           setTemplate.BKBizID,
			SetTemplateName: setTemplate.ServiceInstanceName,
		}
	}

	return result, nil
}

// ListServiceInstanceBySetTemplate list service instance by set template.
func (h *handler) ListServiceInstanceBySetTemplate(ctx context.Context, bizID int64, setTemplate int64,
	page types.Page) ([]*types.ServiceInstance, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListServiceInstanceBySetTemplateReq{
		TenantID: tenantID,

		BKBizID:       bizID,
		SetTemplateID: setTemplate,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}
	resp, err := h.cli.listServiceInstanceBySetTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.ServiceInstance, len(resp.Info))
	for index, serviceInstance := range resp.Info {
		result[index] = &types.ServiceInstance{
			ID:                  serviceInstance.ID,
			BizID:               serviceInstance.BKBizID,
			ServiceInstanceName: serviceInstance.ServiceInstanceName,
			HostID:              serviceInstance.BKHostID,
			ModuleID:            serviceInstance.BKModuleID,
		}
	}

	return result, nil
}

// ListProcTemplate list process template.
func (h *handler) ListProcTemplate(ctx context.Context, bizID int64, serviceTemplateID int64,
	processTemplateIDs []int64) ([]*types.ProcessTemplate, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListProcTemplateReq{
		TenantID:           tenantID,
		BKBizID:            bizID,
		ServiceTemplateID:  serviceTemplateID,
		ProcessTemplateIDs: processTemplateIDs,
	}
	resp, err := h.cli.listProcTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	templates := make([]*types.ProcessTemplate, len(resp.Info))
	for index, procTemplate := range resp.Info {
		templates[index] = &types.ProcessTemplate{
			ID:                  procTemplate.ID,
			BizID:               procTemplate.BKBizID,
			ProcessTemplateName: procTemplate.BKProcessTemplateName,
			ServiceTemplateID:   procTemplate.ServiceTemplateID,
			Property: &types.ProcessProperty{
				AutoStart:       procTemplate.Property.AutoStart.Value,
				BizID:           procTemplate.Property.BKBizID.Value,
				FuncName:        procTemplate.Property.BKFuncName.Value,
				ProcessID:       procTemplate.Property.BKProcessID.Value,
				ProcessName:     procTemplate.Property.BKProcessName.Value,
				StartParamRegex: procTemplate.Property.BKStartParamRegex.Value,
				Description:     procTemplate.Property.Description.Value,
				FaceStopCMD:     procTemplate.Property.FaceStopCMD.Value,
				PidFile:         procTemplate.Property.PidFile.Value,
				Priority:        procTemplate.Property.Priority.Value,
				ProcNum:         procTemplate.Property.ProcNum.Value,
				ReloadCMD:       procTemplate.Property.ReloadCMD.Value,
				RestartCMD:      procTemplate.Property.RestartCMD.Value,
				StartCMD:        procTemplate.Property.StartCMD.Value,
				StopCMD:         procTemplate.Property.StopCMD.Value,
				Timeout:         procTemplate.Property.Timeout.Value,
				User:            procTemplate.Property.User.Value,
				WorkPath:        procTemplate.Property.WorkPath.Value,
			},
		}

		templates[index].Property.BindInfo = make([]*types.BindInfo, len(procTemplate.Property.BindInfo.Value))
		for bindInfoIdx, bindInfo := range procTemplate.Property.BindInfo.Value {
			templates[index].Property.BindInfo[bindInfoIdx] = &types.BindInfo{
				Enable:        bindInfo.Enable.Value,
				IP:            bindInfo.IP.Value,
				Port:          bindInfo.Port.Value,
				Protocol:      bindInfo.Protocol.Value,
				TemplateRowID: bindInfo.TemplateRowID,
			}
		}
	}

	return templates, nil
}

// FindHostBySetTemplate find host by set template.
func (h *handler) FindHostBySetTemplate(ctx context.Context, bizID int64, setTemplateIDs []int64,
	page types.Page) ([]*types.Host, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &FindHostBySetTemplateReq{
		TenantID: tenantID,

		BKBizID:          bizID,
		BKSetTemplateIDs: setTemplateIDs,
		Fields:           DefaultSearchHostFields,
		Page: Page{
			Start: page.Offset,
			Sort:  page.Sort,
			Limit: page.Limit,
		},
	}
	resp, err := h.cli.findHostBySetTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Host, len(resp.Info))
	for index, host := range resp.Info {
		result[index] = &types.Host{
			HostID:   host.BKHostID,
			TenantID: tenantID,
			Static: &types.HostStatic{
				BizID:         bizID,
				NetworkAreaID: host.BKCloudID,
				HostName:      host.BKHostName,
				DeptName:      host.DeptName,
				InnerIP:       host.BKHostInnerIPV4,
				InnerIPV6:     host.BKHostInnerIPV6,
				OuterIP:       host.BKHostOuterIPV4,
				OuterIPV6:     host.BKHostOuterIPV6,
				Mac:           host.BKMac,
				OSType:        host.BKOsType,
				SyncedAgentID: host.BKAgentID,
			},
			Dynamic: types.NewBlankNodeDynamic(),
		}
	}

	return result, nil
}
