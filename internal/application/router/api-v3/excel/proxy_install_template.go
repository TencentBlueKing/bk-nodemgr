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

	columns := proxyInstallColumns()
	for idx := range columns {
		switch columns[idx].Name {
		case colOSType:
			columns[idx].Options = releases.OSType
		case proxyColCPUArch:
			columns[idx].Options = releases.CPUArch
		}
	}

	buffer := new(bytes.Buffer)
	tmpl, err := workbook.NewTemplate(proxyInstallSheetName, columns)
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

// proxyInstallRow carries the raw cell values of one uploaded data row.
type proxyInstallRow struct {
	InnerIP                  string
	InnerIPV6                string
	ExportIP                 string
	Addressing               types.Addressing
	OsType                   criteria.OSType
	LoginIP                  string
	LoginPort                int64
	LoginUser                string
	IsSudo                   bool
	LoginMode                types.LoginMode
	Credit                   string
	CPUArch                  string
	AdvertiseIP              string
	InstallMethod            types.NodeInstallMethod
	InstallPreOrderedPlugins *bool
	ReRegister               *bool
	DedicatedInstaller       *bool
	ClusterTunnel            *bool
	FileTunnel               *bool
	DataTunnel               *bool
}

// nolint: funlen, gocognit, gocyclo, cyclop, maintidx
func proxyInstallColumns() []workbook.TypedColumn[proxyInstallRow] {
	boolOptions := []string{"true", "false"}
	return []workbook.TypedColumn[proxyInstallRow]{
		{
			Column: workbook.Column{
				Name:        colInnerIP,
				Required:    false,
				Description: colDescInnerIP,
			},
			Parse: func(row *proxyInstallRow, v string) error {
				row.InnerIP = v

				return nil
			},
			ValueOf: func(row *proxyInstallRow) any {
				return row.InnerIP
			},
		},
		{
			Column: workbook.Column{
				Name:        colInnerIPV6,
				Required:    false,
				Description: colDescInnerIPV6,
			},
			Parse: func(row *proxyInstallRow, v string) error {
				row.InnerIPV6 = v

				return nil
			},
			ValueOf: func(row *proxyInstallRow) any {
				return row.InnerIPV6
			},
		},
		{
			Column: workbook.Column{
				Name:        proxyColExportIP,
				Required:    true,
				Description: proxyColDescExportIP,
			},
			Parse: func(row *proxyInstallRow, v string) error {
				row.ExportIP = v

				return nil
			},
			ValueOf: func(row *proxyInstallRow) any {
				return row.ExportIP
			},
		},
		{
			Column: workbook.Column{
				Name:        colAddressing,
				Required:    true,
				Description: colDescAddressing,
				Options:     []string{string(types.AddressingStatic), string(types.AddressingDynamic)},
			},
			Parse: func(row *proxyInstallRow, v string) error {
				row.Addressing = types.Addressing(v)
				if err := row.Addressing.Validate(); err != nil {
					return fmt.Errorf("invalid addressing. addressing(%s): %w", v, err)
				}

				return nil
			},
			ValueOf: func(row *proxyInstallRow) any {
				return row.Addressing
			},
		},
		{
			Column: workbook.Column{
				Name:        colOSType,
				Required:    true,
				Description: proxyColDescOSType,
			},
			Parse: func(row *proxyInstallRow, v string) error {
				row.OsType = criteria.OSType(v)
				if err := row.OsType.Validate(); err != nil {
					return fmt.Errorf("invalid os_type. os-type(%s): %w", v, err)
				}

				return nil
			},
			ValueOf: func(row *proxyInstallRow) any {
				return row.OsType
			},
		},
		{
			Column: workbook.Column{
				Name:        colLoginIP,
				Required:    true,
				Description: colDescLoginIP,
			},
			Parse: func(row *proxyInstallRow, v string) error {
				row.LoginIP = v

				return nil
			},
			ValueOf: func(row *proxyInstallRow) any {
				return row.LoginIP
			},
		},
		{
			Column: workbook.Column{
				Name:        colLoginPort,
				Required:    true,
				Description: colDescLoginPort,
			},
			Parse: func(row *proxyInstallRow, v string) error {
				port, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return fmt.Errorf("failed to parse login port. login-port(%s): %w", v, err)
				}
				row.LoginPort = port

				return nil
			},
			ValueOf: func(row *proxyInstallRow) any {
				return row.LoginPort
			},
		},
		{
			Column: workbook.Column{
				Name:        colLoginUser,
				Required:    true,
				Description: colDescLoginUser,
			},
			Parse: func(row *proxyInstallRow, v string) error {
				row.LoginUser = v

				return nil
			},
			ValueOf: func(row *proxyInstallRow) any {
				return row.LoginUser
			},
		},
		{
			Column: workbook.Column{
				Name:        colIsSudo,
				Required:    false,
				Description: colDescIsSudo,
				Options:     boolOptions,
			},
			Parse: func(row *proxyInstallRow, v string) error {
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
			ValueOf: func(row *proxyInstallRow) any {
				return row.IsSudo
			},
		},
		{
			Column: workbook.Column{
				Name:        colLoginMode,
				Required:    true,
				Description: colDescLoginMode,
				Options:     []string{string(types.LoginModePassword), string(types.LoginModeKeyFile), string(types.LoginModePasswordVault)},
			},
			Parse: func(row *proxyInstallRow, v string) error {
				row.LoginMode = types.LoginMode(v)
				if err := row.LoginMode.Validate(); err != nil {
					return fmt.Errorf("invalid login_mode. login-mode(%s): %w", v, err)
				}

				return nil
			},
			ValueOf: func(row *proxyInstallRow) any {
				return row.LoginMode
			},
		},
		{
			Column: workbook.Column{
				Name:        colCredit,
				Required:    true,
				Description: colDescCredit,
			},
			Parse: func(row *proxyInstallRow, v string) error {
				row.Credit = v

				return nil
			},
			ValueOf: func(row *proxyInstallRow) any {
				return row.Credit
			},
		},
		{
			Column: workbook.Column{
				Name:        proxyColCPUArch,
				Required:    false,
				Description: proxyColDescCPUArch,
			},
			Parse: func(row *proxyInstallRow, v string) error {
				row.CPUArch = v

				return nil
			},
			ValueOf: func(row *proxyInstallRow) any {
				return row.CPUArch
			},
		},
		{
			Column: workbook.Column{
				Name:        proxyColAdvertiseIP,
				Required:    false,
				Description: proxyColDescAdvertiseIP,
			},
			Parse: func(row *proxyInstallRow, v string) error {
				row.AdvertiseIP = v

				return nil
			},
			ValueOf: func(row *proxyInstallRow) any {
				return row.AdvertiseIP
			},
		},
		{
			Column: workbook.Column{
				Name:        colInstallPreOrderedPlugins,
				Required:    false,
				Description: proxyColDescInstallPreOrderedPlugins,
				Options:     boolOptions,
			},
			Parse: func(row *proxyInstallRow, v string) error {
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
			ValueOf: func(row *proxyInstallRow) any {
				if row.InstallPreOrderedPlugins == nil {
					return ""
				}

				return *row.InstallPreOrderedPlugins
			},
		},
		{
			Column: workbook.Column{
				Name:        colReRegister,
				Required:    false,
				Description: colDescReRegister,
				Options:     boolOptions,
			},
			Parse: func(row *proxyInstallRow, v string) error {
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
			ValueOf: func(row *proxyInstallRow) any {
				if row.ReRegister == nil {
					return ""
				}

				return *row.ReRegister
			},
		},
		{
			Column: workbook.Column{
				Name:        colInstallMethod,
				Required:    false,
				Description: colDescInstallMethod,
				Options: []string{
					string(types.NodeInstallMethodSSH),
					string(types.NodeInstallMethodWMI),
				},
			},
			Parse: func(row *proxyInstallRow, v string) error {
				row.InstallMethod = types.NodeInstallMethod(v)
				if err := row.InstallMethod.Validate(); err != nil {
					return fmt.Errorf("invalid install_method. value(%s): %w", v, err)
				}

				return nil
			},
			ValueOf: func(row *proxyInstallRow) any {
				return string(row.InstallMethod)
			},
		},
		{
			Column: workbook.Column{
				Name:        proxyColDedicatedInstaller,
				Required:    false,
				Description: proxyColDescDedicatedInstaller,
				Options:     boolOptions,
			},
			Parse: func(row *proxyInstallRow, v string) error {
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
			ValueOf: func(row *proxyInstallRow) any {
				if row.DedicatedInstaller == nil {
					return ""
				}

				return *row.DedicatedInstaller
			},
		},
		{
			Column: workbook.Column{
				Name:        proxyColClusterTunnel,
				Required:    false,
				Description: proxyColDescClusterTunnel,
				Options:     boolOptions,
			},
			Parse: func(row *proxyInstallRow, v string) error {
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
			ValueOf: func(row *proxyInstallRow) any {
				if row.ClusterTunnel == nil {
					return ""
				}

				return *row.ClusterTunnel
			},
		},
		{
			Column: workbook.Column{
				Name:        proxyColFileTunnel,
				Required:    false,
				Description: proxyColDescFileTunnel,
				Options:     boolOptions,
			},
			Parse: func(row *proxyInstallRow, v string) error {
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
			ValueOf: func(row *proxyInstallRow) any {
				if row.FileTunnel == nil {
					return ""
				}

				return *row.FileTunnel
			},
		},
		{
			Column: workbook.Column{
				Name:        proxyColDataTunnel,
				Required:    false,
				Description: proxyColDescDataTunnel,
				Options:     boolOptions,
			},
			Parse: func(row *proxyInstallRow, v string) error {
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
			ValueOf: func(row *proxyInstallRow) any {
				if row.DataTunnel == nil {
					return ""
				}

				return *row.DataTunnel
			},
		},
	}
}

// proxyInstallSampleRows builds the example data rows shown under the template header.
// nolint:mnd
func proxyInstallSampleRows() []proxyInstallRow {
	exampleTrue, exampleFalse := true, false
	exampleCPUArch := "amd64"
	exampleUserDesc := "login_username"
	exampleCreditDesc := "fill in your password or key according to LoginMode"
	var exampleLoginPort int64 = 22
	return []proxyInstallRow{
		{
			InnerIP:                  "1.1.1.1",
			InnerIPV6:                "",
			ExportIP:                 "2.2.2.1",
			Addressing:               types.AddressingStatic,
			OsType:                   criteria.OSLinux,
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
			OsType:                   criteria.OSLinux,
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
			OsType:                   criteria.OSLinux,
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

// ExcelUploadProxyInstallTemplate uploads the proxy install template.
func (h *handler) ExcelUploadProxyInstallTemplate(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ExcelUploadProxyInstallTemplateReq)
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

	tmpl, err := workbook.NewTemplate(proxyInstallSheetName, proxyInstallColumns())
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

		parsedRow.Value.fillInfo(info)
		infos[idx] = info
	}

	resp := new(protoApplication.ExcelUploadProxyInstallTemplateResp)
	resp.Data = &protoApplication.ExcelUploadProxyInstallTemplateResp_Data{Items: infos}

	return resp.GetData(), nil
}

// fillInfo copies one parsed row into the response item.
func (row *proxyInstallRow) fillInfo(info *protoApplication.ExcelProxyInstallTemplateParsedInfo) {
	info.BkHostInnerip = row.InnerIP
	info.BkHostInneripV6 = row.InnerIPV6
	info.ExportIp = row.ExportIP
	info.BkAddressing = string(row.Addressing)
	info.OsType = string(row.OsType)
	info.LoginIp = row.LoginIP
	info.LoginPort = row.LoginPort
	info.LoginUser = row.LoginUser
	info.IsSudo = row.IsSudo
	info.LoginMode = string(row.LoginMode)
	info.Credit = row.Credit
	info.CpuArch = row.CPUArch
	info.AdvertiseIp = row.AdvertiseIP
	info.InstallMethod = string(row.InstallMethod)
	info.InstallPreOrderedPlugins = row.InstallPreOrderedPlugins
	info.ReRegister = row.ReRegister
	info.DedicatedInstaller = row.DedicatedInstaller
	info.ClusterTunnel = row.ClusterTunnel
	info.FileTunnel = row.FileTunnel
	info.DataTunnel = row.DataTunnel
}
