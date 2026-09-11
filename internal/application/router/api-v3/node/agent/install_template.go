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

// Package agent provides the agent API handler.
package agent

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/xuri/excelize/v2"
)

const (
	maxTemplateUploadFileSize = 1024 * 1024 * 10 // 10MB

	sheetName        = "agent_install_template"
	templateFileName = "agent_install_template.xlsx"
)

const (
	colNameInnerIP                  = "内网 IPv4"
	colNameInnerIPV6                = "内网 IPv6（可选，和内网 IPv4 二选一）"
	colNameAddressing               = "寻址方式"
	colNameOSType                   = "操作系统"
	colNameLoginIP                  = "登录IP"
	colNameLoginPort                = "登录端口"
	colNameLoginUser                = "登录用户"
	colNameLoginMode                = "登录方式"
	colNameCredit                   = "密钥/密码"
	colNameInstallPreOrderedPlugins = "安装预设插件"
	colNameReRegister               = "重新注册AgentID"
)

// parsedInfo describes the parsed information from the uploaded agent install template file.
type parsedInfo struct {
	InnerIP                  string
	InnerIPV6                string
	Addressing               types.Addressing
	OsType                   criteria.OSType
	LoginIP                  string
	LoginPort                int64
	LoginUser                string
	LoginMode                types.LoginMode
	Credit                   string
	InstallPreOrderedPlugins *bool
	ReRegister               *bool
}

// column describes one Excel column of the install template.
type column struct {
	name        string
	parse       func(info *parsedInfo, val string) error
	cellValueOf func(info *parsedInfo) any
}

// NOCC: golint/fnsize(template column definitions belong together).
// nolint: funlen, gocognit, gocyclo, cyclop, maintidx
func getColumns() []column {
	return []column{
		{
			name: colNameInnerIP,
			parse: func(info *parsedInfo, v string) error {
				info.InnerIP = v

				return nil
			},
			cellValueOf: func(info *parsedInfo) any {
				return info.InnerIP
			},
		},
		{
			name: colNameInnerIPV6,
			parse: func(info *parsedInfo, v string) error {
				info.InnerIPV6 = v

				return nil
			},
			cellValueOf: func(info *parsedInfo) any {
				return info.InnerIPV6
			},
		},
		{
			name: colNameAddressing,
			parse: func(info *parsedInfo, v string) error {
				info.Addressing = types.Addressing(v)
				if err := info.Addressing.Validate(); err != nil {
					return fmt.Errorf("invalid addressing. addressing(%s): %w", v, err)
				}

				return nil
			},
			cellValueOf: func(info *parsedInfo) any {
				return info.Addressing
			},
		},
		{
			name: colNameOSType,
			parse: func(info *parsedInfo, v string) error {
				info.OsType = criteria.OSType(v)
				if err := info.OsType.Validate(); err != nil {
					return fmt.Errorf("invalid os_type. os-type(%s): %w", v, err)
				}

				return nil
			},
			cellValueOf: func(info *parsedInfo) any {
				return info.OsType
			},
		},
		{
			name: colNameLoginIP,
			parse: func(info *parsedInfo, v string) error {
				info.LoginIP = v

				return nil
			},
			cellValueOf: func(info *parsedInfo) any {
				return info.LoginIP
			},
		},
		{
			name: colNameLoginPort,
			parse: func(info *parsedInfo, v string) error {
				port, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return fmt.Errorf("failed to parse login port. login-port(%s): %w", v, err)
				}
				info.LoginPort = port

				return nil
			},
			cellValueOf: func(info *parsedInfo) any {
				return info.LoginPort
			},
		},
		{
			name: colNameLoginUser,
			parse: func(info *parsedInfo, v string) error {
				info.LoginUser = v

				return nil
			},
			cellValueOf: func(info *parsedInfo) any {
				return info.LoginUser
			},
		},
		{
			name: colNameLoginMode,
			parse: func(info *parsedInfo, v string) error {
				info.LoginMode = types.LoginMode(v)
				if err := info.LoginMode.Validate(); err != nil {
					return fmt.Errorf("invalid login_mode. login-mode(%s): %w", v, err)
				}

				return nil
			},
			cellValueOf: func(info *parsedInfo) any {
				return info.LoginMode
			},
		},
		{
			name: colNameCredit,
			parse: func(info *parsedInfo, v string) error {
				info.Credit = v

				return nil
			},
			cellValueOf: func(info *parsedInfo) any {
				return info.Credit
			},
		},
		{
			name: colNameInstallPreOrderedPlugins,
			parse: func(info *parsedInfo, v string) error {
				if v == "" {
					return nil
				}
				b, err := strconv.ParseBool(v)
				if err != nil {
					return fmt.Errorf("invalid install_pre_ordered_plugins. value(%s): %w", v, err)
				}
				info.InstallPreOrderedPlugins = &b

				return nil
			},
			cellValueOf: func(info *parsedInfo) any {
				if info.InstallPreOrderedPlugins == nil {
					return ""
				}

				return *info.InstallPreOrderedPlugins
			},
		},
		{
			name: colNameReRegister,
			parse: func(info *parsedInfo, v string) error {
				if v == "" {
					return nil
				}
				b, err := strconv.ParseBool(v)
				if err != nil {
					return fmt.Errorf("invalid re_register. value(%s): %w", v, err)
				}
				info.ReRegister = &b

				return nil
			},
			cellValueOf: func(info *parsedInfo) any {
				if info.ReRegister == nil {
					return ""
				}

				return *info.ReRegister
			},
		},
	}
}

// DownloadInstallTemplate downloads the agent install template.
func (h *handler) DownloadInstallTemplate(rCtx restserver.IContext) (*restserver.FileResponse, error) {
	// create template file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to close excel file")
		}
	}()

	if err := f.SetSheetName("Sheet1", sheetName); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to rename default sheet")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	for colIndex, col := range getColumns() {
		cell, _ := excelize.CoordinatesToCellName(colIndex+1, 1)
		if err := f.SetCellValue(sheetName, cell, col.name); err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to set cell value")
			return nil, resterrf.ErrWrap(resterrf.Aborted, err)
		}
	}

	// set sample data
	if err := setSampleData(f); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set sample data")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	buffer := new(bytes.Buffer)
	if err := f.Write(buffer); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to write buffer")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	return &restserver.FileResponse{
		Data:        io.NopCloser(buffer),
		Size:        int64(buffer.Len()),
		FileName:    templateFileName,
		ContentType: restserver.MIMETypeXls,
	}, nil
}

//nolint:mnd
func setSampleData(f *excelize.File) error {
	exampleTrue, exampleFalse := true, false
	exampleUserDesc := "login_username"
	exampleCreditDesc := "fill in your password or key according to LoginMode"
	var exampleLoginPort int64 = 22
	samples := []parsedInfo{
		{
			InnerIP:                  "1.1.1.1",
			InnerIPV6:                "",
			Addressing:               types.AddressingStatic,
			OsType:                   criteria.OSLinux,
			LoginIP:                  "1.1.1.1",
			LoginPort:                exampleLoginPort,
			LoginUser:                exampleUserDesc,
			LoginMode:                types.LoginModePassword,
			Credit:                   exampleCreditDesc,
			InstallPreOrderedPlugins: &exampleTrue,
			ReRegister:               &exampleFalse,
		},
		{
			InnerIP:                  "1.1.1.2",
			InnerIPV6:                "",
			Addressing:               types.AddressingStatic,
			OsType:                   criteria.OSWindows,
			LoginIP:                  "1.1.1.2",
			LoginPort:                exampleLoginPort,
			LoginUser:                exampleUserDesc,
			LoginMode:                types.LoginModeKeyFile,
			Credit:                   exampleCreditDesc,
			InstallPreOrderedPlugins: &exampleTrue,
			ReRegister:               &exampleFalse,
		},
		{
			InnerIP:                  "1.1.1.3",
			InnerIPV6:                "",
			Addressing:               types.AddressingDynamic,
			OsType:                   criteria.OSDarwin,
			LoginIP:                  "1.1.1.3",
			LoginPort:                exampleLoginPort,
			LoginUser:                exampleUserDesc,
			LoginMode:                types.LoginModePasswordVault,
			Credit:                   exampleCreditDesc,
			InstallPreOrderedPlugins: &exampleTrue,
			ReRegister:               &exampleFalse,
		},
	}

	columns := getColumns()
	for rowIndex, info := range samples {
		for colIndex, col := range columns {
			cell, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+2)
			if err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, cell, col.cellValueOf(&info)); err != nil {
				return fmt.Errorf("failed to set sample cell %s: %w", cell, err)
			}
		}
	}

	return nil
}

// UploadInstallTemplate uploads the agent install template.
func (h *handler) UploadInstallTemplate(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.UploadAgentInstallTemplateReq)
	fileHeader, err := rCtx.ParseFileForm(req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload template, failed to parse file form")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if fileHeader.Size > maxTemplateUploadFileSize {
		logger.G.Biz(rCtx).
			With("file_size", fileHeader.Size).
			With("limit", maxTemplateUploadFileSize).
			Error("template file size exceeds limit")

		return nil, resterrf.ErrWrap(resterrf.InvalidFileResource,
			errors.New("template file size exceeds limit"))
	}

	file, err := fileHeader.Open()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload template, failed to open file")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to close excel file")
		}
	}()

	infos, err := parseTemplateToInfos(file)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to parse excel file")
		return nil, resterrf.ErrWrap(resterrf.InvalidFileResource, err)
	}

	return convertParsedInfosToData(infos), nil
}

//nolint:mnd
func parseTemplateToInfos(file io.Reader) ([]parsedInfo, error) {
	f, err := excelize.OpenReader(file, excelize.Options{
		RawCellValue: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer func() { _ = f.Close() }()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, errors.New("no data found in file")
	}

	columns := getColumns()
	if err := validateHeader(rows[0], columns); err != nil {
		return nil, err
	}

	infos := make([]parsedInfo, len(rows)-1)
	for rowIdx, row := range rows[1:] {
		var info parsedInfo
		for colIdx, col := range columns {
			if colIdx >= len(row) {
				break
			}
			if err := col.parse(&info, row[colIdx]); err != nil {
				return nil, fmt.Errorf("failed to parse row. row-index(%d): %w", rowIdx+2, err)
			}
		}
		infos[rowIdx] = info
	}

	return infos, nil
}

func validateHeader(header []string, columns []column) error {
	if len(header) != len(columns) {
		return fmt.Errorf("invalid template header. want(%d) got(%d)", len(columns), len(header))
	}
	for idx, col := range columns {
		if header[idx] != col.name {
			return fmt.Errorf("invalid template header. col(%d) want(%q) got(%q)", idx+1, col.name, header[idx])
		}
	}

	return nil
}

func convertParsedInfosToData(parsedInfos []parsedInfo) *protoApplication.UploadAgentInstallTemplateResp_Data {
	infos := make([]*protoApplication.AgentInstallParsedInfo, len(parsedInfos))
	for idx, info := range parsedInfos {
		infos[idx] = &protoApplication.AgentInstallParsedInfo{
			BkHostInnerip:            info.InnerIP,
			BkHostInneripV6:          info.InnerIPV6,
			BkAddressing:             string(info.Addressing),
			OsType:                   string(info.OsType),
			LoginIp:                  info.LoginIP,
			LoginPort:                info.LoginPort,
			LoginUser:                info.LoginUser,
			LoginMode:                string(info.LoginMode),
			Credit:                   info.Credit,
			InstallPreOrderedPlugins: info.InstallPreOrderedPlugins,
			ReRegister:               info.ReRegister,
		}
	}

	return &protoApplication.UploadAgentInstallTemplateResp_Data{
		Info: infos,
	}
}
