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

package excel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/router/api-v3/excel/workbook"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// opsOutBandProtocol mirrors one entry of the platform-injected out-of-band protocol list.
type opsOutBandProtocol struct {
	Name        string `json:"name"`
	Port        int64  `json:"port"`
	Description string `json:"description"`
}

// ExcelDownloadOpsTemplate renders the ops settings template with live biz/area/protocol options.
func (h *handler) ExcelDownloadOpsTemplate(rCtx restserver.IContext) (*restserver.FileResponse, error) {
	req := new(protoApplication.ExcelDownloadOpsTemplateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download ops template, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	bizs, _, err := h.backendHandler.ListBusiness(rCtx, types.UnlimitedPage(), nil)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list business for ops template options")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}
	bizNames := make([]string, 0, len(bizs))
	for _, biz := range bizs {
		bizNames = append(bizNames, biz.BizName)
	}

	areas, _, err := h.backendHandler.ListNetworkArea(rCtx, types.UnlimitedPage(), nil)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list network areas for ops template options")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}
	areaNames := make([]string, 0, len(areas))
	for _, area := range areas {
		areaNames = append(areaNames, area.Name)
	}

	var protocols []opsOutBandProtocol
	if err := json.Unmarshal([]byte(h.frontSetting.OpsOutBandProtocols()), &protocols); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to parse ops out-of-band protocol options")
	}
	protocolOptions := make([]string, 0, len(protocols))
	for _, protocol := range protocols {
		if protocol.Name != "" {
			protocolOptions = append(protocolOptions, protocol.Name)
		}
	}

	columns := opsTemplateColumns(bizNames, areaNames, protocolOptions)

	buffer := new(bytes.Buffer)
	tmpl, err := workbook.NewTemplate(opsTemplateSheetName, columns, templateLanguage(rCtx))
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create ops template")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}
	if err := tmpl.Write(buffer, opsSampleRows(bizNames, areaNames)); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to render ops template")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	return &restserver.FileResponse{
		Data:        io.NopCloser(buffer),
		Size:        int64(buffer.Len()),
		FileName:    opsTemplateFileName,
		ContentType: restserver.MIMETypeXls,
	}, nil
}

// opsTemplateRow carries the raw cell values of one uploaded data row.
type opsTemplateRow struct {
	InnerIP          string
	InnerIPV6        string
	NetworkArea      string
	Biz              string
	ConsoleArea      string
	ConsoleInnerIP   string
	ConsoleInnerIPV6 string
	OutBandType      string
	OutBandProtocol  string
	OutBandIP        string
	OutBandPort      string
}

// opsTemplateColumns defines the ops template columns.
func opsTemplateColumns(bizNames, areaNames, protocolOptions []string) []workbook.TypedColumn[opsTemplateRow] {
	biz := workbook.TypedColumn[opsTemplateRow]{
		Column: workbook.Column{
			Required: true,
			Options:  bizNames,
		},
		Parse:   func(row *opsTemplateRow, v string) error { row.Biz = v; return nil },
		ValueOf: func(row *opsTemplateRow) any { return row.Biz },
	}
	biz.Name.Edit().Zh(opsColBizZh).En(opsColBizEn).Set()
	biz.Description.Edit().Zh(opsColDescBizZh).En(opsColDescBizEn).Set()

	networkArea := workbook.TypedColumn[opsTemplateRow]{
		Column: workbook.Column{
			Required: true,
			Options:  areaNames,
		},
		Parse:   func(row *opsTemplateRow, v string) error { row.NetworkArea = v; return nil },
		ValueOf: func(row *opsTemplateRow) any { return row.NetworkArea },
	}
	networkArea.Name.Edit().Zh(opsColNetworkAreaZh).En(opsColNetworkAreaEn).Set()
	networkArea.Description.Edit().Zh(opsColDescNetworkAreaZh).En(opsColDescNetworkAreaEn).Set()

	innerIPv4 := workbook.TypedColumn[opsTemplateRow]{
		Column: workbook.Column{
			Required: false,
		},
		Parse:   func(row *opsTemplateRow, v string) error { row.InnerIP = v; return nil },
		ValueOf: func(row *opsTemplateRow) any { return row.InnerIP },
	}
	innerIPv4.Name.Edit().Zh(colInnerIPZh).En(colInnerIPEn).Set()
	innerIPv4.Description.Edit().Zh(opsColDescInnerIPZh).En(opsColDescInnerIPEn).Set()

	innerIPv6 := workbook.TypedColumn[opsTemplateRow]{
		Column: workbook.Column{
			Required: false,
		},
		Parse:   func(row *opsTemplateRow, v string) error { row.InnerIPV6 = v; return nil },
		ValueOf: func(row *opsTemplateRow) any { return row.InnerIPV6 },
	}
	innerIPv6.Name.Edit().Zh(colInnerIPV6Zh).En(colInnerIPV6En).Set()
	innerIPv6.Description.Edit().Zh(opsColDescInnerIPV6Zh).En(opsColDescInnerIPV6En).Set()

	consoleArea := workbook.TypedColumn[opsTemplateRow]{
		Column: workbook.Column{
			Required: false,
			Options:  areaNames,
		},
		Parse:   func(row *opsTemplateRow, v string) error { row.ConsoleArea = v; return nil },
		ValueOf: func(row *opsTemplateRow) any { return row.ConsoleArea },
	}
	consoleArea.Name.Edit().Zh(opsColConsoleAreaZh).En(opsColConsoleAreaEn).Set()
	consoleArea.Description.Edit().Zh(opsColDescConsoleAreaZh).En(opsColDescConsoleAreaEn).Set()

	consoleIPv4 := workbook.TypedColumn[opsTemplateRow]{
		Column: workbook.Column{
			Required: false,
		},
		Parse:   func(row *opsTemplateRow, v string) error { row.ConsoleInnerIP = v; return nil },
		ValueOf: func(row *opsTemplateRow) any { return row.ConsoleInnerIP },
	}
	consoleIPv4.Name.Edit().Zh(opsColConsoleInnerIPZh).En(opsColConsoleInnerIPEn).Set()
	consoleIPv4.Description.Edit().Zh(opsColDescConsoleInnerIPZh).En(opsColDescConsoleInnerIPEn).Set()

	consoleIPv6 := workbook.TypedColumn[opsTemplateRow]{
		Column: workbook.Column{
			Required: false,
		},
		Parse:   func(row *opsTemplateRow, v string) error { row.ConsoleInnerIPV6 = v; return nil },
		ValueOf: func(row *opsTemplateRow) any { return row.ConsoleInnerIPV6 },
	}
	consoleIPv6.Name.Edit().Zh(opsColConsoleInnerIPV6Zh).En(opsColConsoleInnerIPV6En).Set()
	consoleIPv6.Description.Edit().Zh(opsColDescConsoleInnerIPV6Zh).En(opsColDescConsoleInnerIPV6En).Set()

	columns := []workbook.TypedColumn[opsTemplateRow]{biz, networkArea, innerIPv4, innerIPv6, consoleArea, consoleIPv4, consoleIPv6}

	return append(columns, opsOutBandColumns(protocolOptions)...)
}

func opsOutBandColumns(protocolOptions []string) []workbook.TypedColumn[opsTemplateRow] {
	outBandType := workbook.TypedColumn[opsTemplateRow]{
		Column: workbook.Column{
			Required: false,
		},
		Parse:   func(row *opsTemplateRow, v string) error { row.OutBandType = v; return nil },
		ValueOf: func(row *opsTemplateRow) any { return row.OutBandType },
	}
	outBandType.Name.Edit().Zh(opsColOutBandTypeZh).En(opsColOutBandTypeEn).Set()
	outBandType.Description.Edit().Zh(opsColDescOutBandTypeZh).En(opsColDescOutBandTypeEn).Set()

	outBandProtocol := workbook.TypedColumn[opsTemplateRow]{
		Column: workbook.Column{
			Required:    false,
			Options:     protocolOptions,
			AllowCustom: true,
		},
		Parse:   func(row *opsTemplateRow, v string) error { row.OutBandProtocol = v; return nil },
		ValueOf: func(row *opsTemplateRow) any { return row.OutBandProtocol },
	}
	outBandProtocol.Name.Edit().Zh(opsColOutBandProtocolZh).En(opsColOutBandProtocolEn).Set()
	outBandProtocol.Description.Edit().Zh(opsColDescOutBandProtocolZh).En(opsColDescOutBandProtocolEn).Set()

	outBandIP := workbook.TypedColumn[opsTemplateRow]{
		Column: workbook.Column{
			Required: false,
		},
		Parse:   func(row *opsTemplateRow, v string) error { row.OutBandIP = v; return nil },
		ValueOf: func(row *opsTemplateRow) any { return row.OutBandIP },
	}
	outBandIP.Name.Edit().Zh(opsColOutBandIPZh).En(opsColOutBandIPEn).Set()
	outBandIP.Description.Edit().Zh(opsColDescOutBandIPZh).En(opsColDescOutBandIPEn).Set()

	outBandPort := workbook.TypedColumn[opsTemplateRow]{
		Column: workbook.Column{
			Required: false,
		},
		Parse:   func(row *opsTemplateRow, v string) error { row.OutBandPort = v; return nil },
		ValueOf: func(row *opsTemplateRow) any { return row.OutBandPort },
	}
	outBandPort.Name.Edit().Zh(opsColOutBandPortZh).En(opsColOutBandPortEn).Set()
	outBandPort.Description.Edit().Zh(opsColDescOutBandPortZh).En(opsColDescOutBandPortEn).Set()

	return []workbook.TypedColumn[opsTemplateRow]{outBandType, outBandProtocol, outBandIP, outBandPort}
}

// opsSampleRows builds example rows using real dropdown names when available.
func opsSampleRows(bizNames, areaNames []string) []opsTemplateRow {
	bizName, areaName := "", ""
	if len(bizNames) > 0 {
		bizName = bizNames[0]
	}
	if len(areaNames) > 0 {
		areaName = areaNames[0]
	}

	return []opsTemplateRow{
		{
			InnerIP:         "1.1.1.1",
			NetworkArea:     areaName,
			Biz:             bizName,
			OutBandType:     "onboard",
			OutBandProtocol: "ipmi",
			OutBandIP:       "10.0.0.1",
			OutBandPort:     "623",
		},
	}
}

// ExcelParseOpsTemplate parses the uploaded ops template and resolves each row against existing hosts.
func (h *handler) ExcelParseOpsTemplate(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ExcelParseOpsTemplateReq)
	fileHeader, err := rCtx.ParseFileForm(req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to parse ops template upload form")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if fileHeader.Size > maxTemplateUploadFileSize {
		logger.G.Biz(rCtx).
			With("file_size", fileHeader.Size).
			With("limit", maxTemplateUploadFileSize).
			Error("ops template file size exceeds limit")

		return nil, resterrf.ErrWrap(resterrf.InvalidFileResource,
			errors.New("template file size exceeds limit"))
	}
	file, err := fileHeader.Open()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to open ops template file")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to close ops template file")
		}
	}()

	tmpl, err := workbook.NewTemplate(opsTemplateSheetName, opsTemplateColumns(nil, nil, nil), workbook.LanguageZh)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create ops template")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	rows, err := tmpl.ParseIndexed(file)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to parse ops template file")
		return nil, resterrf.ErrWrap(resterrf.InvalidFileResource, err)
	}

	validRows := slices.Clone(rows)
	validRows = slices.DeleteFunc(validRows, func(row workbook.IndexedRow[opsTemplateRow]) bool {
		return row.Err != nil
	})

	resolver, err := h.newOpsRowResolver(rCtx, validRows)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to resolve ops template rows")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	infos := make([]*protoApplication.ExcelOpsTemplateParsedInfo, len(rows))
	for idx, parsedRow := range rows {
		info := &protoApplication.ExcelOpsTemplateParsedInfo{RowIndex: int64(parsedRow.RowIndex)}
		infos[idx] = info
		if parsedRow.Err != nil {
			info.ErrorMessage = parsedRow.Err.Error()

			continue
		}

		result, err := resolver.resolve(&parsedRow.Value)
		if err != nil {
			info.ErrorMessage = err.Error()

			continue
		}
		info.ConvertFromTypes(&result)
	}

	resp := new(protoApplication.ExcelParseOpsTemplateResp)
	resp.Data = &protoApplication.ExcelParseOpsTemplateResp_Data{Items: infos}

	return resp.GetData(), nil
}

// opsRowResolver holds the preloaded lookup tables shared by all rows.
type opsRowResolver struct {
	bizIDs       map[string]int64
	areaIDs      map[string]int64
	targetHosts  []*types.Host
	consoleHosts []*types.Host
}

// newOpsRowResolver preloads biz/area name maps and batch-queries
// target/console hosts.
func (h *handler) newOpsRowResolver(
	rCtx restserver.IContext,
	rows []workbook.IndexedRow[opsTemplateRow],
) (*opsRowResolver, error) {

	bizs, _, err := h.backendHandler.ListBusiness(rCtx, types.UnlimitedPage(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list business: %w", err)
	}
	areas, _, err := h.backendHandler.ListNetworkArea(rCtx, types.UnlimitedPage(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list network areas: %w", err)
	}

	resolver := &opsRowResolver{
		bizIDs:  make(map[string]int64, len(bizs)),
		areaIDs: make(map[string]int64, len(areas)),
	}
	for _, biz := range bizs {
		if _, exists := resolver.bizIDs[biz.BizName]; exists {
			resolver.bizIDs[biz.BizName] = -1
			continue
		}
		resolver.bizIDs[biz.BizName] = biz.BizID
	}
	for _, area := range areas {
		if _, exists := resolver.areaIDs[area.Name]; exists {
			resolver.areaIDs[area.Name] = -1
			continue
		}
		resolver.areaIDs[area.Name] = area.ID
	}

	var targetIPs, targetIPV6s, consoleIPs, consoleIPV6s []string
	for _, indexed := range rows {
		row := indexed.Value
		if row.InnerIP != "" {
			targetIPs = append(targetIPs, row.InnerIP)
		}
		if row.InnerIPV6 != "" {
			targetIPV6s = append(targetIPV6s, row.InnerIPV6)
		}
		if row.ConsoleInnerIP != "" {
			consoleIPs = append(consoleIPs, row.ConsoleInnerIP)
		}
		if row.ConsoleInnerIPV6 != "" {
			consoleIPV6s = append(consoleIPV6s, row.ConsoleInnerIPV6)
		}
	}

	resolver.targetHosts, err = h.listHostsByIPs(rCtx, targetIPs, targetIPV6s)
	if err != nil {
		return nil, fmt.Errorf("failed to list target hosts: %w", err)
	}
	resolver.consoleHosts, err = h.listHostsByIPs(rCtx, consoleIPs, consoleIPV6s)
	if err != nil {
		return nil, fmt.Errorf("failed to list console hosts: %w", err)
	}

	return resolver, nil
}

// listHostsByIPs batch-queries hosts whose inner IPv4 or IPv6 matches any given IP.
func (h *handler) listHostsByIPs(rCtx restserver.IContext, ips, ipv6s []string) ([]*types.Host, error) {
	if len(ips) == 0 && len(ipv6s) == 0 {
		return nil, nil
	}

	merged := make([]*types.Host, 0)
	seen := make(map[int64]struct{})
	appendHosts := func(hosts []*types.Host) {
		for _, host := range hosts {
			if _, exists := seen[host.HostID]; exists {
				continue
			}
			seen[host.HostID] = struct{}{}
			merged = append(merged, host)
		}
	}

	if len(ips) > 0 {
		hosts, _, err := h.backendHandler.ListHost(rCtx, types.UnlimitedPage(), &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{InnerIP: ips},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list hosts by innerip: %w", err)
		}
		appendHosts(hosts)
	}

	if len(ipv6s) > 0 {
		hosts, _, err := h.backendHandler.ListHost(rCtx, types.UnlimitedPage(), &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{InnerIPV6: ipv6s},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list hosts by innerip v6: %w", err)
		}
		appendHosts(hosts)
	}

	return merged, nil
}

// resolve validates an Ops row and returns the matched hosts and settings.
func (resolver *opsRowResolver) resolve(row *opsTemplateRow) (types.ExcelOpsTemplateRowResult, error) {
	var result types.ExcelOpsTemplateRowResult
	if row.InnerIP == "" && row.InnerIPV6 == "" {
		return result, errors.New("inner IPv4 and inner IPv6 can not both be empty")
	}

	bizID, err := resolver.lookupName(resolver.bizIDs, row.Biz, "business")
	if err != nil {
		return result, err
	}
	areaID, err := resolver.lookupName(resolver.areaIDs, row.NetworkArea, "network area")
	if err != nil {
		return result, err
	}

	host, err := matchHost(resolver.targetHosts, bizID, areaID, row.InnerIP, row.InnerIPV6)
	if err != nil {
		return result, fmt.Errorf("failed to match target host: %w", err)
	}

	var port int64
	if row.OutBandPort != "" {
		port, err = strconv.ParseInt(row.OutBandPort, 10, 64)
		if err != nil {
			return result, fmt.Errorf("invalid out-of-band port: %s", row.OutBandPort)
		}
	}

	var consoleHost *types.Host
	if row.ConsoleArea != "" || row.ConsoleInnerIP != "" || row.ConsoleInnerIPV6 != "" {
		consoleHost, err = resolver.resolveConsoleHost(row)
		if err != nil {
			return result, err
		}
	}

	result = types.ExcelOpsTemplateRowResult{
		HostID:             host.HostID,
		BizID:              bizID,
		BizName:            row.Biz,
		NetworkAreaID:      areaID,
		NetworkAreaName:    row.NetworkArea,
		InnerIP:            row.InnerIP,
		InnerIPV6:          row.InnerIPV6,
		OpsOutBandType:     row.OutBandType,
		OpsOutBandProtocol: row.OutBandProtocol,
		OpsOutBandIP:       row.OutBandIP,
		OpsOutBandPort:     port,
	}
	if consoleHost == nil {
		return result, nil
	}

	result.OpsConsoleHostID = consoleHost.HostID
	if len(consoleHost.Static.InnerIPList) > 0 {
		result.OpsConsoleHostIP = consoleHost.Static.InnerIPList[0]
	} else if len(consoleHost.Static.InnerIPV6List) > 0 {
		result.OpsConsoleHostIP = consoleHost.Static.InnerIPV6List[0]
	}

	return result, nil
}

// resolveConsoleHost matches a configured console host by its network area and IP.
func (resolver *opsRowResolver) resolveConsoleHost(row *opsTemplateRow) (*types.Host, error) {
	if row.ConsoleInnerIP == "" && row.ConsoleInnerIPV6 == "" {
		return nil, errors.New("console IPv4 and console IPv6 can not both be empty")
	}

	consoleAreaID, err := resolver.lookupName(resolver.areaIDs, row.ConsoleArea, "console network area")
	if err != nil {
		return nil, err
	}

	host, err := matchHost(resolver.consoleHosts, -1, consoleAreaID, row.ConsoleInnerIP, row.ConsoleInnerIPV6)
	if err != nil {
		return nil, fmt.Errorf("failed to match console host: %w", err)
	}

	return host, nil
}

// lookupName resolves a display name to its ID, rejecting unknown or ambiguous names.
func (resolver *opsRowResolver) lookupName(names map[string]int64, name, field string) (int64, error) {
	if name == "" {
		return 0, fmt.Errorf("%s can not be empty", field)
	}
	id, exists := names[name]
	if !exists {
		return 0, fmt.Errorf("%s not found: %s", field, name)
	}
	if id < 0 {
		return 0, fmt.Errorf("%s name is ambiguous: %s", field, name)
	}

	return id, nil
}

// matchHost finds exactly one host under the biz/area scope matching either IP.
func matchHost(hosts []*types.Host, bizID, areaID int64, innerIP, innerIPV6 string) (*types.Host, error) {
	var matched *types.Host
	for _, host := range hosts {
		if host.Static == nil {
			continue
		}
		if bizID >= 0 && host.Static.BizID != bizID {
			continue
		}
		if host.Static.NetworkAreaID != areaID {
			continue
		}
		ipMatched := (innerIP != "" && slices.Contains(host.Static.InnerIPList, innerIP)) ||
			(innerIPV6 != "" && slices.Contains(host.Static.InnerIPV6List, innerIPV6))
		if !ipMatched {
			continue
		}
		if matched != nil {
			return nil, errors.New("matched multiple hosts, check IP and network area")
		}
		matched = host
	}
	if matched == nil {
		return nil, errors.New("no matching host found")
	}

	return matched, nil
}
