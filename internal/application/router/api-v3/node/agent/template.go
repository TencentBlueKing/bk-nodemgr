/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package agent provides the agent API handler.
package agent

import (
	"bytes"
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
	sheetName        = "install_template"
	templateFileName = "install_template.xlsx"

	exampleUserDesc   = "login_username"
	exampleCreditDesc = "fill in your password or key according to LoginMode"
	exampleLoginPort  = 22
)

type column struct {
	key  string
	name string
}

const (
	templateKeyInnerIP   = "bk_host_innerip"
	templateKeyInnerIPV6 = "bk_host_inneripv6"
	templateKeyOsType    = "os_type"
	templateKeyLoginIP   = "login_ip"
	templateKeyLoginPort = "login_port"
	templateKeyLoginUser = "login_user"
	templateKeyLoginMode = "login_mode"
	templateKeyCredit    = "credit"

	templateNameInnerIP   = "内网 IPv4"
	templateNameInnerIPV6 = "内网 IPv6"
	templateNameOsType    = "操作系统"
	templateNameLoginIP   = "登录IP"
	templateNameLoginPort = "登录端口"
	templateNameLoginUser = "登录用户"
	templateNameLoginMode = "登陆方式"
	templateNameCredit    = "密钥/密码"
)

func getColumns() []column {
	return []column{
		{key: templateKeyInnerIP, name: templateNameInnerIP},
		{key: templateKeyInnerIPV6, name: templateNameInnerIPV6},
		{key: templateKeyOsType, name: templateNameOsType},
		{key: templateKeyLoginIP, name: templateNameLoginIP},
		{key: templateKeyLoginPort, name: templateNameLoginPort},
		{key: templateKeyLoginUser, name: templateNameLoginUser},
		{key: templateKeyLoginMode, name: templateNameLoginMode},
		{key: templateKeyCredit, name: templateNameCredit},
	}
}

// DownloadTemplate downloads the agent install template.
func (h *handler) DownloadTemplate(rCtx restserver.IContext) (*restserver.FileResponse, error) {
	// create template file
	f := excelize.NewFile()
	defer f.Close()

	// create sheet
	index, err := f.NewSheet(sheetName)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create sheet")
		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to create sheet: %w", err))
	}

	// set active sheet
	f.SetActiveSheet(index)

	// delete default sheet
	if err := f.DeleteSheet("Sheet1"); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete default sheet")
		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to delete default sheet: %w", err))
	}

	// set header
	for colIndex, col := range getColumns() {
		cell, _ := excelize.CoordinatesToCellName(colIndex+1, 1)
		if err := f.SetCellValue(sheetName, cell, col.name); err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to set cell value")
			return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to set cell value: %w", err))
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
		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to write buffer: %w", err))
	}

	return &restserver.FileResponse{
		Data:        io.NopCloser(buffer),
		Size:        int64(buffer.Len()),
		FileName:    templateFileName,
		ContentType: restserver.MIMETypeXls,
	}, nil
}

// setSampleData sets sample data to the excel file.
// nolint: errcheck, mnd
func setSampleData(f *excelize.File) error {
	sampleInfos := []parsedInfo{
		{
			InnerIP:   "1.1.1.1",
			InnerIPV6: "",
			OsType:    criteria.OSLinux,
			LoginIP:   "1.1.1.1",
			LoginPort: exampleLoginPort,
			LoginUser: exampleUserDesc,
			LoginMode: types.LoginModePassword,
			Credit:    exampleCreditDesc,
		},
		{
			InnerIP:   "1.1.1.2",
			InnerIPV6: "",
			OsType:    criteria.OSWindows,
			LoginIP:   "1.1.1.2",
			LoginPort: exampleLoginPort,
			LoginUser: exampleUserDesc,
			LoginMode: types.LoginModeKeyFile,
			Credit:    exampleCreditDesc,
		},
		{
			InnerIP:   "1.1.1.3",
			InnerIPV6: "",
			OsType:    criteria.OSDarwin,
			LoginIP:   "1.1.1.3",
			LoginPort: exampleLoginPort,
			LoginUser: exampleUserDesc,
			LoginMode: types.LoginModePasswordVault,
			Credit:    exampleCreditDesc,
		},
	}

	for rowIndex, info := range sampleInfos {
		rowData := info.toRowData()
		for colIndex, cellData := range rowData {
			cell, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+2)
			if err != nil {
				return err
			}
			f.SetCellValue(sheetName, cell, cellData)
		}
	}

	return nil
}

// UploadTemplate parses the uploaded template and returns the data.
func (h *handler) UploadTemplate(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.UploadAgentTemplateReq)
	fileHeader, err := rCtx.ParseFileForm(req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload template, failed to parse file form")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload template, failed to open file")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer file.Close()

	// parse excel file to infos
	infos, err := parseExcelToInfos(file)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to parse excel file")
		return nil, resterrf.ErrWrap(resterrf.InvalidFileResource, err)
	}

	respData := convertParsedInfosToData(infos)

	return respData, nil
}

// parseExcelToInfos parses the uploaded excel file and returns parsed info slice.
// nolint: errcheck, mnd
func parseExcelToInfos(file io.Reader) ([]parsedInfo, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("no data found in file")
	}

	columns := getColumns()
	infos := make([]parsedInfo, len(rows)-1)
	for idx, row := range rows[1:] {
		rowMap := make(map[string]string)
		for colIndex, cell := range row {
			rowMap[columns[colIndex].key] = cell
		}

		info, err := parseRowToInfo(rowMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse row: %w", err)
		}
		infos[idx] = info
	}

	return infos, nil
}

// parseRowToInfo parses a row map to parsedInfo.
// nolint: errcheck,unparam
func parseRowToInfo(rowMap map[string]string) (parsedInfo, error) {
	info := parsedInfo{}

	parsers := map[string]func(string) error{
		templateKeyInnerIP:   func(val string) error { info.InnerIP = val; return nil },
		templateKeyInnerIPV6: func(val string) error { info.InnerIPV6 = val; return nil },
		templateKeyOsType: func(val string) error {
			info.OsType = criteria.OSType(val)
			if err := info.OsType.Validate(); err != nil {
				return fmt.Errorf("failed to validate os type. os-type(%s):%w", val, err)
			}

			return nil
		},
		templateKeyLoginIP: func(val string) error { info.LoginIP = val; return nil },
		templateKeyLoginPort: func(val string) error {
			port, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse login port. port(%s)", val)
			}
			info.LoginPort = port

			return nil
		},
		templateKeyLoginUser: func(val string) error { info.LoginUser = val; return nil },
		templateKeyLoginMode: func(val string) error {
			info.LoginMode = types.LoginMode(val)
			if err := info.LoginMode.Validate(); err != nil {
				return fmt.Errorf("failed to validate login mode. mode(%s):%w", val, err)
			}

			return nil
		},
		templateKeyCredit: func(val string) error { info.Credit = val; return nil },
	}

	for key, parseFunc := range parsers {
		val, exists := rowMap[key]
		if !exists {
			continue
		}

		if err := parseFunc(val); err != nil {
			return info, err
		}
	}

	return info, nil
}

// parsedInfo describes the parsed information from the uploaded template file.
type parsedInfo struct {
	InnerIP   string
	InnerIPV6 string
	OsType    criteria.OSType
	LoginIP   string
	LoginPort int64
	LoginUser string
	LoginMode types.LoginMode
	Credit    string
}

func (info *parsedInfo) toRowData() []any {
	return []any{
		info.InnerIP,
		info.InnerIPV6,
		info.OsType,
		info.LoginIP,
		info.LoginPort,
		info.LoginUser,
		info.LoginMode,
		info.Credit,
	}
}

// convertParsedInfosToData converts parsedInfo slice to UploadAgentTemplateResp_Data.
func convertParsedInfosToData(parsedInfos []parsedInfo) *protoApplication.UploadAgentTemplateResp_Data {
	if len(parsedInfos) == 0 {
		return &protoApplication.UploadAgentTemplateResp_Data{
			Info:       []*protoApplication.ParsedInfo{},
			TotalCount: 0,
		}
	}

	infos := make([]*protoApplication.ParsedInfo, len(parsedInfos))
	for idx, info := range parsedInfos {
		infos[idx] = &protoApplication.ParsedInfo{
			InnerIp:    info.InnerIP,
			InnerIpv6:  info.InnerIPV6,
			OsType:     string(info.OsType),
			LoginIp:    info.LoginIP,
			LoginPort:  info.LoginPort,
			LoginUser:  info.LoginUser,
			LoginMode:  string(info.LoginMode),
			Credential: info.Credit,
		}
	}

	return &protoApplication.UploadAgentTemplateResp_Data{
		Info:       infos,
		TotalCount: int64(len(parsedInfos)),
	}
}
