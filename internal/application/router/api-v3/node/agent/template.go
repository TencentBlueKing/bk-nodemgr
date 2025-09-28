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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/xuri/excelize/v2"
)

const (
	sheetName = "install_template"
	fileName  = "install_template.xlsx"
)

type column struct {
	key  string
	name string
}

const (
	templateKeyInnerIP   = "inner_ip"
	templateKeyInnerIPV6 = "inner_ipv6"
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
	data, err := createTemplate()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create template: %v", err)
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	return &restserver.FileResponse{
		Data:        data,
		FileName:    fileName,
		ContentType: restserver.MIMETypeXls,
	}, nil
}

// createTemplate create agent install template.
// nolint:errcheck
func createTemplate() (io.ReadCloser, error) {
	f := excelize.NewFile()
	defer f.Close()

	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("create sheet failed: %w", err)
	}

	f.SetActiveSheet(index)
	// delete default sheet
	f.DeleteSheet("Sheet1")

	// set header
	for colIndex, col := range getColumns() {
		cell, _ := excelize.CoordinatesToCellName(colIndex+1, 1)
		f.SetCellValue(sheetName, cell, col.name)
	}

	// set sample data
	if err := setSampleData(f); err != nil {
		return nil, err
	}

	buffer := new(bytes.Buffer)
	if err := f.Write(buffer); err != nil {
		return nil, fmt.Errorf("write to buffer failed: %w", err)
	}

	return io.NopCloser(buffer), nil
}

// setSampleData sets sample data to the excel file.
// nolint: errcheck, mnd
func setSampleData(f *excelize.File) error {
	sampleInfos := []types.ParsedInfo{
		{
			InnerIP:   "1.1.1.1",
			InnerIPV6: "",
			OsType:    "linux",
			LoginIP:   "1.1.1.1",
			LoginPort: 22,
			LoginUser: "root",
			LoginMode: "password",
			Credit:    "123456",
		},
		{
			InnerIP:   "1.1.1.2",
			InnerIPV6: "",
			OsType:    "windows",
			LoginIP:   "1.1.1.2",
			LoginPort: 36000,
			LoginUser: "Administrator",
			LoginMode: "password",
			Credit:    "66666",
		},
		{
			InnerIP:   "1.1.1.3",
			InnerIPV6: "",
			OsType:    "linux",
			LoginIP:   "1.1.1.3",
			LoginPort: 36000,
			LoginUser: "root",
			LoginMode: "password",
			Credit:    "8888888",
		},
	}

	for rowIndex, info := range sampleInfos {
		rowData := info.ToRowData()
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
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload template, failed to parse file form: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload template, failed to open file: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer file.Close()

	// parse excel file to infos
	infos, err := parseExcelToInfos(file)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to parse excel file: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidFileResource, err)
	}

	logger.G.Biz(rCtx).With("info-count", len(infos)).Info("successfully parsed uploaded template")

	resp := new(protoApplication.UploadAgentTemplateResp)
	resp.ConvertResultFromData(infos)

	return resp.GetData(), nil
}

// parseExcelToInfos parses the uploaded excel file and returns parsed info slice.
// nolint: errcheck, mnd
func parseExcelToInfos(file io.Reader) ([]types.ParsedInfo, error) {
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
	infos := make([]types.ParsedInfo, len(rows)-1)
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

// parseRowToInfo parses a row map to ParsedInfo.
// nolint: errcheck,unparam
func parseRowToInfo(rowMap map[string]string) (types.ParsedInfo, error) {
	info := types.ParsedInfo{}

	parsers := map[string]func(string) error{
		templateKeyInnerIP:   func(val string) error { info.InnerIP = val; return nil },
		templateKeyInnerIPV6: func(val string) error { info.InnerIPV6 = val; return nil },
		templateKeyOsType:    func(val string) error { info.OsType = val; return nil },
		templateKeyLoginIP:   func(val string) error { info.LoginIP = val; return nil },
		templateKeyLoginPort: func(val string) error {
			port, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse login port. port(%s)", val)
			}
			info.LoginPort = port

			return nil
		},
		templateKeyLoginUser: func(val string) error { info.LoginUser = val; return nil },
		templateKeyLoginMode: func(val string) error { info.LoginMode = val; return nil },
		templateKeyCredit:    func(val string) error { info.Credit = val; return nil },
	}

	for key, parseFunc := range parsers {
		if val, exists := rowMap[key]; exists {
			if err := parseFunc(val); err != nil {
				return info, err
			}
		}
	}

	return info, nil
}
