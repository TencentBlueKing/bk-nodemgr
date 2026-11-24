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

// ExcelDownloadAgentInstallTemplate downloads the agent install template.
func (h *handler) ExcelDownloadAgentInstallTemplate(rCtx restserver.IContext) (*restserver.FileResponse, error) {
	req := new(protoApplication.ExcelDownloadAgentInstallTemplateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download agent install template, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	releases, err := h.backendHandler.DistinctReleaseAgent(rCtx, types.Generation2, types.ReleaseDistinctField{OSType: true},
		&types.ReleaseCondition{ExactInclude: &types.ReleaseExactFields{Enabled: []bool{true}}})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct agent releases for template options")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	columns := agentInstallColumns(releases.OSType)

	buffer := new(bytes.Buffer)
	tmpl, err := workbook.NewTemplate(agentInstallSheetName, columns, templateLanguage(rCtx))
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create agent install template")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}
	if err := tmpl.Write(buffer, agentInstallSampleRows()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to render agent install template")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	return &restserver.FileResponse{
		Data:        io.NopCloser(buffer),
		Size:        int64(buffer.Len()),
		FileName:    agentInstallFileName,
		ContentType: restserver.MIMETypeXls,
	}, nil
}

func agentInstallColumns(osTypes []string) []workbook.TypedColumn[types.ExcelAgentInstallTemplateRow] {
	boolOptions := []string{"true", "false"}
	columns := agentInstallHostColumns(osTypes)
	columns = append(columns, agentInstallLoginColumns(boolOptions)...)

	return append(columns, agentInstallOptionColumns(boolOptions)...)
}

func agentInstallHostColumns(osTypes []string) []workbook.TypedColumn[types.ExcelAgentInstallTemplateRow] {
	innerIPv4 := workbook.TypedColumn[types.ExcelAgentInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
		},
		Parse: func(row *types.ExcelAgentInstallTemplateRow, v string) error {
			row.InnerIP = v

			return nil
		},
		ValueOf: func(row *types.ExcelAgentInstallTemplateRow) any {
			return row.InnerIP
		},
	}
	innerIPv4.Name.Edit().Zh(colInnerIPZh).En(colInnerIPEn).Set()
	innerIPv4.Description.Edit().Zh(colDescInnerIPZh).En(colDescInnerIPEn).Set()

	innerIPv6 := workbook.TypedColumn[types.ExcelAgentInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
		},
		Parse: func(row *types.ExcelAgentInstallTemplateRow, v string) error {
			row.InnerIPV6 = v

			return nil
		},
		ValueOf: func(row *types.ExcelAgentInstallTemplateRow) any {
			return row.InnerIPV6
		},
	}
	innerIPv6.Name.Edit().Zh(colInnerIPV6Zh).En(colInnerIPV6En).Set()
	innerIPv6.Description.Edit().Zh(colDescInnerIPV6Zh).En(colDescInnerIPV6En).Set()

	addressing := workbook.TypedColumn[types.ExcelAgentInstallTemplateRow]{
		Column: workbook.Column{
			Required: true,
			Options:  []string{string(types.AddressingStatic), string(types.AddressingDynamic)},
		},
		Parse: func(row *types.ExcelAgentInstallTemplateRow, v string) error {
			row.Addressing = types.Addressing(v)
			if err := row.Addressing.Validate(); err != nil {
				return fmt.Errorf("invalid addressing. addressing(%s): %w", v, err)
			}

			return nil
		},
		ValueOf: func(row *types.ExcelAgentInstallTemplateRow) any {
			return row.Addressing
		},
	}
	addressing.Name.Edit().Zh(colAddressingZh).En(colAddressingEn).Set()
	addressing.Description.Edit().Zh(colDescAddressingZh).En(colDescAddressingEn).Set()

	osType := workbook.TypedColumn[types.ExcelAgentInstallTemplateRow]{
		Column: workbook.Column{
			Required: true,
			Options:  osTypes,
		},
		Parse: func(row *types.ExcelAgentInstallTemplateRow, v string) error {
			row.OSType = criteria.OSType(v)
			if err := row.OSType.Validate(); err != nil {
				return fmt.Errorf("invalid os_type. os-type(%s): %w", v, err)
			}

			return nil
		},
		ValueOf: func(row *types.ExcelAgentInstallTemplateRow) any {
			return row.OSType
		},
	}
	osType.Name.Edit().Zh(colOSTypeZh).En(colOSTypeEn).Set()
	osType.Description.Edit().Zh(agentColDescOSTypeZh).En(agentColDescOSTypeEn).Set()

	return []workbook.TypedColumn[types.ExcelAgentInstallTemplateRow]{innerIPv4, innerIPv6, addressing, osType}
}

func agentInstallLoginColumns(boolOptions []string) []workbook.TypedColumn[types.ExcelAgentInstallTemplateRow] {
	loginIP := workbook.TypedColumn[types.ExcelAgentInstallTemplateRow]{
		Column: workbook.Column{
			Required: true,
		},
		Parse: func(row *types.ExcelAgentInstallTemplateRow, v string) error {
			row.LoginIP = v

			return nil
		},
		ValueOf: func(row *types.ExcelAgentInstallTemplateRow) any {
			return row.LoginIP
		},
	}
	loginIP.Name.Edit().Zh(colLoginIPZh).En(colLoginIPEn).Set()
	loginIP.Description.Edit().Zh(colDescLoginIPZh).En(colDescLoginIPEn).Set()

	loginPort := workbook.TypedColumn[types.ExcelAgentInstallTemplateRow]{
		Column: workbook.Column{
			Required: true,
		},
		Parse: func(row *types.ExcelAgentInstallTemplateRow, v string) error {
			port, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse login port. login-port(%s): %w", v, err)
			}
			row.LoginPort = port

			return nil
		},
		ValueOf: func(row *types.ExcelAgentInstallTemplateRow) any {
			return row.LoginPort
		},
	}
	loginPort.Name.Edit().Zh(colLoginPortZh).En(colLoginPortEn).Set()
	loginPort.Description.Edit().Zh(colDescLoginPortZh).En(colDescLoginPortEn).Set()

	loginUser := workbook.TypedColumn[types.ExcelAgentInstallTemplateRow]{
		Column: workbook.Column{
			Required: true,
		},
		Parse: func(row *types.ExcelAgentInstallTemplateRow, v string) error {
			row.LoginUser = v

			return nil
		},
		ValueOf: func(row *types.ExcelAgentInstallTemplateRow) any {
			return row.LoginUser
		},
	}
	loginUser.Name.Edit().Zh(colLoginUserZh).En(colLoginUserEn).Set()
	loginUser.Description.Edit().Zh(colDescLoginUserZh).En(colDescLoginUserEn).Set()

	isSudo := workbook.TypedColumn[types.ExcelAgentInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
			Options:  boolOptions,
		},
		Parse: func(row *types.ExcelAgentInstallTemplateRow, v string) error {
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
		ValueOf: func(row *types.ExcelAgentInstallTemplateRow) any {
			return row.IsSudo
		},
	}
	isSudo.Name.Edit().Zh(colIsSudoZh).En(colIsSudoEn).Set()
	isSudo.Description.Edit().Zh(colDescIsSudoZh).En(colDescIsSudoEn).Set()

	loginMode := workbook.TypedColumn[types.ExcelAgentInstallTemplateRow]{
		Column: workbook.Column{
			Required: true,
			Options:  []string{string(types.LoginModePassword), string(types.LoginModeKeyFile), string(types.LoginModePasswordVault)},
		},
		Parse: func(row *types.ExcelAgentInstallTemplateRow, v string) error {
			row.LoginMode = types.LoginMode(v)
			if err := row.LoginMode.Validate(); err != nil {
				return fmt.Errorf("invalid login_mode. login-mode(%s): %w", v, err)
			}

			return nil
		},
		ValueOf: func(row *types.ExcelAgentInstallTemplateRow) any {
			return row.LoginMode
		},
	}
	loginMode.Name.Edit().Zh(colLoginModeZh).En(colLoginModeEn).Set()
	loginMode.Description.Edit().Zh(colDescLoginModeZh).En(colDescLoginModeEn).Set()

	credit := workbook.TypedColumn[types.ExcelAgentInstallTemplateRow]{
		Column: workbook.Column{
			Required: true,
		},
		Parse: func(row *types.ExcelAgentInstallTemplateRow, v string) error {
			row.Credit = v

			return nil
		},
		ValueOf: func(row *types.ExcelAgentInstallTemplateRow) any {
			return row.Credit
		},
	}
	credit.Name.Edit().Zh(colCreditZh).En(colCreditEn).Set()
	credit.Description.Edit().Zh(colDescCreditZh).En(colDescCreditEn).Set()

	return []workbook.TypedColumn[types.ExcelAgentInstallTemplateRow]{loginIP, loginPort, loginUser, isSudo, loginMode, credit}
}

func agentInstallOptionColumns(boolOptions []string) []workbook.TypedColumn[types.ExcelAgentInstallTemplateRow] {
	installMethod := workbook.TypedColumn[types.ExcelAgentInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
			Options: []string{
				string(types.NodeInstallMethodSSH),
				string(types.NodeInstallMethodWMI),
			},
		},
		Parse: func(row *types.ExcelAgentInstallTemplateRow, v string) error {
			row.InstallMethod = types.NodeInstallMethod(v)
			if err := row.InstallMethod.Validate(); err != nil {
				return fmt.Errorf("invalid install_method. value(%s): %w", v, err)
			}

			return nil
		},
		ValueOf: func(row *types.ExcelAgentInstallTemplateRow) any {
			return string(row.InstallMethod)
		},
	}
	installMethod.Name.Edit().Zh(colInstallMethodZh).En(colInstallMethodEn).Set()
	installMethod.Description.Edit().Zh(colDescInstallMethodZh).En(colDescInstallMethodEn).Set()

	installPreOrderedPlugins := workbook.TypedColumn[types.ExcelAgentInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
			Options:  boolOptions,
		},
		Parse: func(row *types.ExcelAgentInstallTemplateRow, v string) error {
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
		ValueOf: func(row *types.ExcelAgentInstallTemplateRow) any {
			if row.InstallPreOrderedPlugins == nil {
				return ""
			}

			return *row.InstallPreOrderedPlugins
		},
	}
	installPreOrderedPlugins.Name.Edit().Zh(colInstallPreOrderedPluginsZh).En(colInstallPreOrderedPluginsEn).Set()
	installPreOrderedPlugins.Description.Edit().Zh(agentColDescInstallPreOrderedPluginsZh).En(agentColDescInstallPreOrderedPluginsEn).Set()

	reRegister := workbook.TypedColumn[types.ExcelAgentInstallTemplateRow]{
		Column: workbook.Column{
			Required: false,
			Options:  boolOptions,
		},
		Parse: func(row *types.ExcelAgentInstallTemplateRow, v string) error {
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
		ValueOf: func(row *types.ExcelAgentInstallTemplateRow) any {
			if row.ReRegister == nil {
				return ""
			}

			return *row.ReRegister
		},
	}
	reRegister.Name.Edit().Zh(colReRegisterZh).En(colReRegisterEn).Set()
	reRegister.Description.Edit().Zh(colDescReRegisterZh).En(colDescReRegisterEn).Set()

	return []workbook.TypedColumn[types.ExcelAgentInstallTemplateRow]{installMethod, installPreOrderedPlugins, reRegister}
}

// agentInstallSampleRows builds the example data rows shown under the template header.
//
//nolint:mnd
func agentInstallSampleRows() []types.ExcelAgentInstallTemplateRow {
	exampleTrue, exampleFalse := true, false
	exampleUserDesc := "login_username"
	exampleCreditDesc := "fill in your password or key according to LoginMode"
	var exampleLoginPort int64 = 22
	return []types.ExcelAgentInstallTemplateRow{
		{
			InnerIP:                  "1.1.1.1",
			InnerIPV6:                "",
			Addressing:               types.AddressingStatic,
			OSType:                   criteria.OSLinux,
			LoginIP:                  "1.1.1.1",
			LoginPort:                exampleLoginPort,
			LoginUser:                exampleUserDesc,
			IsSudo:                   false,
			LoginMode:                types.LoginModePassword,
			Credit:                   exampleCreditDesc,
			InstallPreOrderedPlugins: &exampleTrue,
			ReRegister:               &exampleFalse,
			InstallMethod:            types.NodeInstallMethodSSH,
		},
		{
			InnerIP:                  "1.1.1.2",
			InnerIPV6:                "",
			Addressing:               types.AddressingStatic,
			OSType:                   criteria.OSWindows,
			LoginIP:                  "1.1.1.2",
			LoginPort:                exampleLoginPort,
			LoginUser:                exampleUserDesc,
			IsSudo:                   false,
			LoginMode:                types.LoginModeKeyFile,
			Credit:                   exampleCreditDesc,
			InstallPreOrderedPlugins: &exampleTrue,
			ReRegister:               &exampleFalse,
			InstallMethod:            types.NodeInstallMethodSSH,
		},
		{
			InnerIP:                  "1.1.1.3",
			InnerIPV6:                "",
			Addressing:               types.AddressingDynamic,
			OSType:                   criteria.OSDarwin,
			LoginIP:                  "1.1.1.3",
			LoginPort:                exampleLoginPort,
			LoginUser:                exampleUserDesc,
			IsSudo:                   false,
			LoginMode:                types.LoginModePasswordVault,
			Credit:                   exampleCreditDesc,
			InstallPreOrderedPlugins: &exampleTrue,
			ReRegister:               &exampleFalse,
			InstallMethod:            types.NodeInstallMethodSSH,
		},
	}
}

// ExcelParseAgentInstallTemplate parses the agent install template.
func (h *handler) ExcelParseAgentInstallTemplate(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ExcelParseAgentInstallTemplateReq)
	fileHeader, err := rCtx.ParseFileForm(req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to parse agent install template upload form")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if fileHeader.Size > maxTemplateUploadFileSize {
		logger.G.Biz(rCtx).
			With("file_size", fileHeader.Size).
			With("limit", maxTemplateUploadFileSize).
			Error("agent install template file size exceeds limit")

		return nil, resterrf.ErrWrap(resterrf.InvalidFileResource,
			errors.New("template file size exceeds limit"))
	}
	file, err := fileHeader.Open()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to open agent install template file")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to close agent install template file")
		}
	}()

	tmpl, err := workbook.NewTemplate(agentInstallSheetName, agentInstallColumns(nil), workbook.LanguageZh)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create agent install template")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	rows, err := tmpl.ParseIndexed(file)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to parse agent install template file")
		return nil, resterrf.ErrWrap(resterrf.InvalidFileResource, err)
	}

	infos := make([]*protoApplication.ExcelAgentInstallTemplateParsedInfo, len(rows))
	for idx, parsedRow := range rows {
		info := &protoApplication.ExcelAgentInstallTemplateParsedInfo{RowIndex: int64(parsedRow.RowIndex)}
		if parsedRow.Err != nil {
			info.ErrorMessage = parsedRow.Err.Error()
			infos[idx] = info

			continue
		}

		info.ConvertFromTypes(&parsedRow.Value)
		infos[idx] = info
	}

	resp := new(protoApplication.ExcelParseAgentInstallTemplateResp)
	resp.Data = &protoApplication.ExcelParseAgentInstallTemplateResp_Data{Items: infos}

	return resp.GetData(), nil
}
