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
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/router/api-v3/excel/workbook"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ExcelDownloadProxyInstallTemplate downloads the proxy install template.
func (h *handler) ExcelDownloadProxyInstallTemplate(rCtx restserver.IContext) (*restserver.FileResponse, error) {
	req := new(protoApplication.ExcelDownloadProxyInstallTemplateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download proxy install template, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	releases, err := h.backendHandler.DistinctReleaseProxy(rCtx, types.Generation2,
		types.ReleaseDistinctField{OSType: true, CPUArch: true},
		&types.ReleaseCondition{ExactInclude: &types.ReleaseExactFields{Enabled: []bool{true}}})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct proxy releases for template options")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	columns := proxyInstallColumns(releases.OSType, releases.CPUArch)

	buffer := new(bytes.Buffer)
	tmpl, err := workbook.NewTemplate(proxyInstallSheetName, columns, templateLanguage(rCtx))
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create proxy install template")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}
	if err := tmpl.Write(buffer, proxyInstallSampleRows()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to render proxy install template")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	return &restserver.FileResponse{
		Data:        io.NopCloser(buffer),
		Size:        int64(buffer.Len()),
		FileName:    proxyInstallFileName,
		ContentType: restserver.MIMETypeXls,
	}, nil
}

func proxyInstallColumns(osTypes, cpuArchs []string) []workbook.TypedColumn[types.ExcelProxyInstallTemplateRow] {
	boolOptions := []string{"true", "false"}
	columns := proxyInstallAddressColumns()
	columns = append(columns, proxyInstallHostColumns(osTypes, cpuArchs)...)
	columns = append(columns, proxyInstallLoginColumns(boolOptions)...)
	columns = append(columns, proxyInstallOptionColumns(boolOptions)...)

	return append(columns, proxyInstallTunnelColumns(boolOptions)...)
}

func proxyInstallAddressColumns() []workbook.TypedColumn[types.ExcelProxyInstallTemplateRow] {
	innerIPv4 := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			row.InnerIP = v

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			return row.InnerIP
		},
	}
	innerIPv4.Name.Edit().Zh(colInnerIPZh).En(colInnerIPEn).Set()
	innerIPv4.Description.Edit().Zh(colDescInnerIPZh).En(colDescInnerIPEn).Set()

	innerIPv6 := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			row.InnerIPV6 = v

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			return row.InnerIPV6
		},
	}
	innerIPv6.Name.Edit().Zh(colInnerIPV6Zh).En(colInnerIPV6En).Set()
	innerIPv6.Description.Edit().Zh(colDescInnerIPV6Zh).En(colDescInnerIPV6En).Set()

	exportIP := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: true,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			row.ExportIP = v

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			return row.ExportIP
		},
	}
	exportIP.Name.Edit().Zh(proxyColExportIPZh).En(proxyColExportIPEn).Set()
	exportIP.Description.Edit().Zh(proxyColDescExportIPZh).En(proxyColDescExportIPEn).Set()

	addressing := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: true,
			Options:  []string{string(types.AddressingStatic), string(types.AddressingDynamic)},
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			row.Addressing = types.Addressing(v)
			if err := row.Addressing.Validate(); err != nil {
				return fmt.Errorf("invalid addressing. addressing(%s): %w", v, err)
			}

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			return row.Addressing
		},
	}
	addressing.Name.Edit().Zh(colAddressingZh).En(colAddressingEn).Set()
	addressing.Description.Edit().Zh(colDescAddressingZh).En(colDescAddressingEn).Set()

	return []workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{innerIPv4, innerIPv6, exportIP, addressing}
}

func proxyInstallHostColumns(osTypes, cpuArchs []string) []workbook.TypedColumn[types.ExcelProxyInstallTemplateRow] {
	osType := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: true,
			Options:  osTypes,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			row.OSType = criteria.OSType(v)
			if err := row.OSType.Validate(); err != nil {
				return fmt.Errorf("invalid os_type. os-type(%s): %w", v, err)
			}

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			return row.OSType
		},
	}
	osType.Name.Edit().Zh(colOSTypeZh).En(colOSTypeEn).Set()
	osType.Description.Edit().Zh(proxyColDescOSTypeZh).En(proxyColDescOSTypeEn).Set()

	cpuArch := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
			Options:  cpuArchs,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			row.CPUArch = v

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			return row.CPUArch
		},
	}
	cpuArch.Name.Edit().Zh(proxyColCPUArchZh).En(proxyColCPUArchEn).Set()
	cpuArch.Description.Edit().Zh(proxyColDescCPUArchZh).En(proxyColDescCPUArchEn).Set()

	advertiseIP := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			row.AdvertiseIP = v

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			return row.AdvertiseIP
		},
	}
	advertiseIP.Name.Edit().Zh(proxyColAdvertiseIPZh).En(proxyColAdvertiseIPEn).Set()
	advertiseIP.Description.Edit().Zh(proxyColDescAdvertiseIPZh).En(proxyColDescAdvertiseIPEn).Set()

	return []workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{osType, cpuArch, advertiseIP}
}

func proxyInstallLoginColumns(boolOptions []string) []workbook.TypedColumn[types.ExcelProxyInstallTemplateRow] {
	loginIP := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: true,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			row.LoginIP = v

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			return row.LoginIP
		},
	}
	loginIP.Name.Edit().Zh(colLoginIPZh).En(colLoginIPEn).Set()
	loginIP.Description.Edit().Zh(colDescLoginIPZh).En(colDescLoginIPEn).Set()

	loginPort := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: true,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			port, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse login port. login-port(%s): %w", v, err)
			}
			row.LoginPort = port

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			return row.LoginPort
		},
	}
	loginPort.Name.Edit().Zh(colLoginPortZh).En(colLoginPortEn).Set()
	loginPort.Description.Edit().Zh(colDescLoginPortZh).En(colDescLoginPortEn).Set()

	loginUser := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: true,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			row.LoginUser = v

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			return row.LoginUser
		},
	}
	loginUser.Name.Edit().Zh(colLoginUserZh).En(colLoginUserEn).Set()
	loginUser.Description.Edit().Zh(colDescLoginUserZh).En(colDescLoginUserEn).Set()

	isSudo := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
			Options:  boolOptions,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			if v == "" {
				return nil
			}
			value, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("invalid is_sudo. value(%s): %w", v, err)
			}
			row.IsSudo = value

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			return row.IsSudo
		},
	}
	isSudo.Name.Edit().Zh(colIsSudoZh).En(colIsSudoEn).Set()
	isSudo.Description.Edit().Zh(colDescIsSudoZh).En(colDescIsSudoEn).Set()

	loginMode := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: true,
			Options:  []string{string(types.LoginModePassword), string(types.LoginModeKeyFile), string(types.LoginModePasswordVault)},
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			row.LoginMode = types.LoginMode(v)
			if err := row.LoginMode.Validate(); err != nil {
				return fmt.Errorf("invalid login_mode. login-mode(%s): %w", v, err)
			}

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			return row.LoginMode
		},
	}
	loginMode.Name.Edit().Zh(colLoginModeZh).En(colLoginModeEn).Set()
	loginMode.Description.Edit().Zh(colDescLoginModeZh).En(colDescLoginModeEn).Set()

	credit := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: true,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			row.Credit = v

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			return row.Credit
		},
	}
	credit.Name.Edit().Zh(colCreditZh).En(colCreditEn).Set()
	credit.Description.Edit().Zh(colDescCreditZh).En(colDescCreditEn).Set()

	return []workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{loginIP, loginPort, loginUser, isSudo, loginMode, credit}
}

func proxyInstallOptionColumns(boolOptions []string) []workbook.TypedColumn[types.ExcelProxyInstallTemplateRow] {
	installMethod := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
			Options: []string{
				string(types.NodeInstallMethodSSH),
				string(types.NodeInstallMethodWMI),
			},
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			row.InstallMethod = types.NodeInstallMethod(v)
			if err := row.InstallMethod.Validate(); err != nil {
				return fmt.Errorf("invalid install_method. value(%s): %w", v, err)
			}

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			return string(row.InstallMethod)
		},
	}
	installMethod.Name.Edit().Zh(colInstallMethodZh).En(colInstallMethodEn).Set()
	installMethod.Description.Edit().Zh(colDescInstallMethodZh).En(colDescInstallMethodEn).Set()

	installPreOrderedPlugins := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
			Options:  boolOptions,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			if v == "" {
				return nil
			}
			value, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("invalid install_pre_ordered_plugins. value(%s): %w", v, err)
			}
			row.InstallPreOrderedPlugins = &value

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			if row.InstallPreOrderedPlugins == nil {
				return ""
			}

			return *row.InstallPreOrderedPlugins
		},
	}
	installPreOrderedPlugins.Name.Edit().Zh(colInstallPreOrderedPluginsZh).En(colInstallPreOrderedPluginsEn).Set()
	installPreOrderedPlugins.Description.Edit().Zh(proxyColDescInstallPreOrderedPluginsZh).En(proxyColDescInstallPreOrderedPluginsEn).Set()

	reRegister := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
			Options:  boolOptions,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			if v == "" {
				return nil
			}
			value, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("invalid re_register. value(%s): %w", v, err)
			}
			row.ReRegister = &value

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			if row.ReRegister == nil {
				return ""
			}

			return *row.ReRegister
		},
	}
	reRegister.Name.Edit().Zh(colReRegisterZh).En(colReRegisterEn).Set()
	reRegister.Description.Edit().Zh(colDescReRegisterZh).En(colDescReRegisterEn).Set()

	dedicatedInstaller := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
			Options:  boolOptions,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			if v == "" {
				return nil
			}
			value, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("invalid dedicated_installer. value(%s): %w", v, err)
			}
			row.DedicatedInstaller = &value

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			if row.DedicatedInstaller == nil {
				return ""
			}

			return *row.DedicatedInstaller
		},
	}
	dedicatedInstaller.Name.Edit().Zh(proxyColDedicatedInstallerZh).En(proxyColDedicatedInstallerEn).Set()
	dedicatedInstaller.Description.Edit().Zh(proxyColDescDedicatedInstallerZh).En(proxyColDescDedicatedInstallerEn).Set()

	return []workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{installMethod, installPreOrderedPlugins, reRegister, dedicatedInstaller}
}

func proxyInstallTunnelColumns(boolOptions []string) []workbook.TypedColumn[types.ExcelProxyInstallTemplateRow] {
	clusterTunnel := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
			Options:  boolOptions,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			if v == "" {
				return nil
			}
			value, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("invalid cluster_tunnel. value(%s): %w", v, err)
			}
			row.ClusterTunnel = &value

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			if row.ClusterTunnel == nil {
				return ""
			}

			return *row.ClusterTunnel
		},
	}
	clusterTunnel.Name.Edit().Zh(proxyColClusterTunnelZh).En(proxyColClusterTunnelEn).Set()
	clusterTunnel.Description.Edit().Zh(proxyColDescClusterTunnelZh).En(proxyColDescClusterTunnelEn).Set()

	fileTunnel := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
			Options:  boolOptions,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			if v == "" {
				return nil
			}
			value, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("invalid file_tunnel. value(%s): %w", v, err)
			}
			row.FileTunnel = &value

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			if row.FileTunnel == nil {
				return ""
			}

			return *row.FileTunnel
		},
	}
	fileTunnel.Name.Edit().Zh(proxyColFileTunnelZh).En(proxyColFileTunnelEn).Set()
	fileTunnel.Description.Edit().Zh(proxyColDescFileTunnelZh).En(proxyColDescFileTunnelEn).Set()

	dataTunnel := workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
			Options:  boolOptions,
		},
		Parse: func(row *types.ExcelProxyInstallTemplateRow, v string) error {
			if v == "" {
				return nil
			}
			value, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("invalid data_tunnel. value(%s): %w", v, err)
			}
			row.DataTunnel = &value

			return nil
		},
		ValueOf: func(row *types.ExcelProxyInstallTemplateRow) any {
			if row.DataTunnel == nil {
				return ""
			}

			return *row.DataTunnel
		},
	}
	dataTunnel.Name.Edit().Zh(proxyColDataTunnelZh).En(proxyColDataTunnelEn).Set()
	dataTunnel.Description.Edit().Zh(proxyColDescDataTunnelZh).En(proxyColDescDataTunnelEn).Set()

	return []workbook.TypedColumn[types.ExcelProxyInstallTemplateRow]{clusterTunnel, fileTunnel, dataTunnel}
}

// proxyInstallSampleRows builds the example data rows shown under the template header.
// nolint:mnd
func proxyInstallSampleRows() []types.ExcelProxyInstallTemplateRow {
	exampleTrue, exampleFalse := true, false
	exampleCPUArch := "amd64"
	exampleUserDesc := "login_username"
	exampleCreditDesc := "fill in your password or key according to LoginMode"
	var exampleLoginPort int64 = 22
	return []types.ExcelProxyInstallTemplateRow{
		{
			InnerIP:                  "1.1.1.1",
			InnerIPV6:                "",
			ExportIP:                 "2.2.2.1",
			Addressing:               types.AddressingStatic,
			OSType:                   criteria.OSLinux,
			LoginIP:                  "1.1.1.1",
			LoginPort:                exampleLoginPort,
			LoginUser:                exampleUserDesc,
			IsSudo:                   false,
			LoginMode:                types.LoginModePassword,
			Credit:                   exampleCreditDesc,
			CPUArch:                  exampleCPUArch,
			AdvertiseIP:              "3.3.3.1",
			InstallMethod:            types.NodeInstallMethodSSH,
			InstallPreOrderedPlugins: &exampleTrue,
			ReRegister:               &exampleFalse,
			DedicatedInstaller:       &exampleTrue,
			ClusterTunnel:            &exampleTrue,
			FileTunnel:               &exampleTrue,
			DataTunnel:               &exampleTrue,
		},
		{
			InnerIP:                  "1.1.1.2",
			InnerIPV6:                "",
			ExportIP:                 "2.2.2.2",
			Addressing:               types.AddressingStatic,
			OSType:                   criteria.OSLinux,
			LoginIP:                  "1.1.1.2",
			LoginPort:                exampleLoginPort,
			LoginUser:                exampleUserDesc,
			IsSudo:                   false,
			LoginMode:                types.LoginModeKeyFile,
			Credit:                   exampleCreditDesc,
			CPUArch:                  exampleCPUArch,
			AdvertiseIP:              "3.3.3.2",
			InstallMethod:            types.NodeInstallMethodSSH,
			InstallPreOrderedPlugins: &exampleTrue,
			ReRegister:               &exampleFalse,
			DedicatedInstaller:       &exampleTrue,
			ClusterTunnel:            &exampleTrue,
			FileTunnel:               &exampleTrue,
			DataTunnel:               &exampleTrue,
		},
		{
			InnerIP:                  "1.1.1.3",
			InnerIPV6:                "",
			ExportIP:                 "2.2.2.3",
			Addressing:               types.AddressingDynamic,
			OSType:                   criteria.OSLinux,
			LoginIP:                  "1.1.1.3",
			LoginPort:                exampleLoginPort,
			LoginUser:                exampleUserDesc,
			IsSudo:                   false,
			LoginMode:                types.LoginModePasswordVault,
			Credit:                   exampleCreditDesc,
			CPUArch:                  exampleCPUArch,
			AdvertiseIP:              "3.3.3.3",
			InstallMethod:            types.NodeInstallMethodSSH,
			InstallPreOrderedPlugins: &exampleTrue,
			ReRegister:               &exampleFalse,
			DedicatedInstaller:       &exampleTrue,
			ClusterTunnel:            &exampleTrue,
			FileTunnel:               &exampleTrue,
			DataTunnel:               &exampleTrue,
		},
	}
}

// ExcelParseProxyInstallTemplate parses the proxy install template.
func (h *handler) ExcelParseProxyInstallTemplate(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ExcelParseProxyInstallTemplateReq)
	fileHeader, err := rCtx.ParseFileForm(req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to parse proxy install template upload form")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if fileHeader.Size > maxTemplateUploadFileSize {
		logger.G.Biz(rCtx).
			With("file_size", fileHeader.Size).
			With("limit", maxTemplateUploadFileSize).
			Error("proxy install template file size exceeds limit")

		return nil, resterrf.ErrWrap(resterrf.InvalidFileResource,
			errors.New("template file size exceeds limit"))
	}
	file, err := fileHeader.Open()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to open proxy install template file")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to close proxy install template file")
		}
	}()

	tmpl, err := workbook.NewTemplate(proxyInstallSheetName, proxyInstallColumns(nil, nil), workbook.LanguageZh)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create proxy install template")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	rows, err := tmpl.ParseIndexed(file)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to parse proxy install template file")
		return nil, resterrf.ErrWrap(resterrf.InvalidFileResource, err)
	}

	infos := make([]*protoApplication.ExcelProxyInstallTemplateParsedInfo, len(rows))
	for idx, parsedRow := range rows {
		info := &protoApplication.ExcelProxyInstallTemplateParsedInfo{RowIndex: int64(parsedRow.RowIndex)}
		if parsedRow.Err != nil {
			info.ErrorMessage = parsedRow.Err.Error()
			infos[idx] = info

			continue
		}

		info.ConvertFromTypes(&parsedRow.Value)
		infos[idx] = info
	}

	resp := new(protoApplication.ExcelParseProxyInstallTemplateResp)
	resp.Data = &protoApplication.ExcelParseProxyInstallTemplateResp_Data{Items: infos}

	return resp.GetData(), nil
}
