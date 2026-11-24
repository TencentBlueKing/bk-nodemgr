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

package workbook

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func testColumns() []Column {
	var osDescription Text
	osDescription.Edit().Zh("目标主机操作系统类型。").En("Operating system of the %s.", "target host").Set()

	return []Column{
		{
			Name:        testText("内网 IPv4", "Internal IPv4"),
			Description: testText("目标主机内网 IPv4 地址，与「内网 IPv6」至少填写一个。", "Enter at least one of Internal IPv4 and Internal IPv6."),
		},
		{
			Name: testText("操作系统", "Operating System"), Required: true,
			Description: osDescription,
			Options:     []string{"linux", "windows"},
		},
		{Name: testText("备注", "Note"), Description: testText("自由填写。", "Enter any text.")},
	}
}

func testText(zh, en string) Text {
	var text Text
	text.Edit().Zh(zh).En(en).Set()

	return text
}

func renderTemplate(t *testing.T, columns []Column, samples [][]any, language Language) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	if err := writeTemplate(buf, "test_sheet", columns, samples, language); err != nil {
		t.Fatalf("writeTemplate failed: %v", err)
	}
	return buf.Bytes()
}

func TestWriteTemplateAndParseRowsRoundTrip(t *testing.T) {
	columns := testColumns()
	samples := [][]any{
		{"1.1.1.1", "linux", "first"},
		{"1.1.1.2", "windows", "second"},
	}
	data := renderTemplate(t, columns, samples, LanguageZh)

	rows, _, err := parseRows(bytes.NewReader(data), "test_sheet", columns)
	if err != nil {
		t.Fatalf("parseRows failed: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 data rows, got %d", len(rows))
	}
	if rows[0][0] != "1.1.1.1" || rows[0][1] != "linux" || rows[1][2] != "second" {
		t.Fatalf("unexpected rows: %v", rows)
	}
}

func TestWriteTemplateLayoutAndDropdowns(t *testing.T) {
	columns := testColumns()
	data := renderTemplate(t, columns, [][]any{{"1.1.1.1", "linux", ""}}, LanguageEn)

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	rows, err := f.GetRows("test_sheet")
	if err != nil {
		t.Fatal(err)
	}
	if rows[0][1] != "Operating System" || rows[1][1] != "Required" || rows[1][0] != "Optional" ||
		rows[2][1] != "Operating system of the target host." {
		t.Fatalf("unexpected header rows: %v", rows[:3])
	}

	dvs, err := f.GetDataValidations("test_sheet")
	if err != nil {
		t.Fatal(err)
	}
	if len(dvs) != 1 {
		t.Fatalf("want 1 dropdown, got %d", len(dvs))
	}
	if dvs[0].Sqref != fmt.Sprintf("B4:B%d", excelize.TotalRows) {
		t.Fatalf("dropdown should cover whole column, got %s", dvs[0].Sqref)
	}
	if dvs[0].Formula1 != `"linux,windows"` {
		t.Fatalf("unexpected inline formula: %s", dvs[0].Formula1)
	}
	if dvs[0].ErrorTitle == nil || *dvs[0].ErrorTitle != "Invalid Input" ||
		dvs[0].Error == nil || *dvs[0].Error != "Select a value from the drop-down list." {
		t.Fatalf("unexpected drop-down notice: %+v", dvs[0])
	}

	for row := 1; row <= 3; row++ {
		height, err := f.GetRowHeight("test_sheet", row)
		if err != nil {
			t.Fatal(err)
		}
		if height <= 0 {
			t.Fatalf("row %d should have explicit height", row)
		}
	}
}

func TestWriteTemplateLongOptionsUseHiddenSheet(t *testing.T) {
	options := make([]string, 0, 60)
	for idx := 0; idx < 60; idx++ {
		options = append(options, fmt.Sprintf("业务名称-%02d", idx))
	}
	columns := []Column{
		{
			Name: testText("业务", "Business"), Required: true,
			Description: testText("目标主机归属业务。", "Business that owns the target host."),
			Options:     options,
		},
	}
	data := renderTemplate(t, columns, [][]any{{options[0]}}, LanguageEn)

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	dvs, err := f.GetDataValidations("test_sheet")
	if err != nil {
		t.Fatal(err)
	}
	if len(dvs) != 1 {
		t.Fatalf("want 1 dropdown, got %d", len(dvs))
	}
	if !strings.Contains(dvs[0].Formula1, "_options_1") {
		t.Fatalf("long list should reference hidden sheet, got %s", dvs[0].Formula1)
	}

	visible, err := f.GetSheetVisible("_options_1")
	if err != nil {
		t.Fatal(err)
	}
	if visible {
		t.Fatal("options sheet should be hidden")
	}

	rows, _, err := parseRows(bytes.NewReader(data), "test_sheet", columns)
	if err != nil {
		t.Fatalf("parseRows failed: %v", err)
	}
	if len(rows) != 1 || rows[0][0] != options[0] {
		t.Fatalf("unexpected rows: %v", rows)
	}
}

func TestWriteTemplateAllowCustomUsesWarning(t *testing.T) {
	columns := []Column{
		{
			Name:        testText("设备协议", "Device Protocol"),
			Description: testText("带外协议。", "Out-of-band protocol."),
			Options:     []string{"ipmi"}, AllowCustom: true,
		},
	}
	data := renderTemplate(t, columns, nil, LanguageEn)

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	dvs, err := f.GetDataValidations("test_sheet")
	if err != nil {
		t.Fatal(err)
	}
	if len(dvs) != 1 || dvs[0].ErrorStyle == nil || *dvs[0].ErrorStyle != "warning" {
		t.Fatalf("allow-custom dropdown should use warning style: %+v", dvs)
	}
	if dvs[0].ErrorTitle == nil || *dvs[0].ErrorTitle != "Input Guidance" ||
		dvs[0].Error == nil || *dvs[0].Error != "Select a value from the drop-down list or enter a custom value." {
		t.Fatalf("unexpected drop-down notice: %+v", dvs[0])
	}
}

func TestWriteTemplateRequiredFlagControlsStyle(t *testing.T) {
	columns := []Column{
		{Name: testText("必填列", "Required Column"), Required: true, Description: testText("必填列说明。", "Required field.")},
		{Name: testText("选填列", "Optional Column"), Description: testText("选填列说明。", "Optional field.")},
	}
	data := renderTemplate(t, columns, nil, LanguageZh)

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	rows, err := f.GetRows("test_sheet")
	if err != nil {
		t.Fatal(err)
	}
	if rows[1][0] != "必填" || rows[1][1] != "可选" {
		t.Fatalf("required flag should control the second row label, got %v", rows[1])
	}

	requiredStyle, err := f.GetCellStyle("test_sheet", "A2")
	if err != nil {
		t.Fatal(err)
	}
	optionalStyle, err := f.GetCellStyle("test_sheet", "B2")
	if err != nil {
		t.Fatal(err)
	}
	if requiredStyle == optionalStyle {
		t.Fatal("optional flag should control the required row style")
	}
}

func TestParseRowsRejectsLegacyLayout(t *testing.T) {
	columns := testColumns()

	f := excelize.NewFile()
	if err := f.SetSheetName("Sheet1", "test_sheet"); err != nil {
		t.Fatal(err)
	}
	for colIndex, col := range columns {
		cell, _ := excelize.CoordinatesToCellName(colIndex+1, 1)
		if err := f.SetCellValue("test_sheet", cell, col.Name.in(LanguageZh)); err != nil {
			t.Fatal(err)
		}
	}
	for row := 2; row <= 4; row++ {
		cell, _ := excelize.CoordinatesToCellName(1, row)
		if err := f.SetCellValue("test_sheet", cell, fmt.Sprintf("1.1.1.%d", row)); err != nil {
			t.Fatal(err)
		}
	}
	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	_, _, err := parseRows(bytes.NewReader(buf.Bytes()), "test_sheet", columns)
	requireErrorContaining(t, err, "outdated excel template layout", "legacy layout should be rejected")
}

func TestParseRowsValidatesHeader(t *testing.T) {
	const (
		invalidHeader  = "invalid excel template header, please download the latest template"
		outdatedLayout = "outdated excel template layout, please download the latest template"
	)
	columns := testColumns()
	data := renderTemplate(t, columns, [][]any{{"1.1.1.1", "linux", ""}}, LanguageZh)

	for _, testCase := range []struct {
		name  string
		cells map[string]string
		want  string
	}{
		{name: "changed title", cells: map[string]string{"B1": "被改掉的表头"}, want: invalidHeader},
		{name: "mixed titles and labels", cells: map[string]string{"B1": "Operating System", "A2": "Optional"}},
		{name: "reordered columns", cells: map[string]string{"A1": "操作系统", "B1": "内网 IPv4"}, want: invalidHeader},
		{name: "extra column", cells: map[string]string{"D1": "额外列"}, want: invalidHeader},
		{name: "mismatched label", cells: map[string]string{"B2": "Optional"}, want: outdatedLayout},
		{
			name: "English labels under Chinese titles", cells: map[string]string{"A2": "Optional", "B2": "Required", "C2": "Optional"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			f, err := excelize.OpenReader(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = f.Close() }()
			for cell, value := range testCase.cells {
				if err := f.SetCellValue("test_sheet", cell, value); err != nil {
					t.Fatal(err)
				}
			}
			buf := new(bytes.Buffer)
			if err := f.Write(buf); err != nil {
				t.Fatal(err)
			}

			_, _, err = parseRows(bytes.NewReader(buf.Bytes()), "test_sheet", columns)
			if testCase.want == "" && err == nil {
				return
			}
			if got := fmt.Sprint(err); got != testCase.want {
				t.Fatalf("parseRows error = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestParseRowsNoData(t *testing.T) {
	columns := testColumns()
	data := renderTemplate(t, columns, nil, LanguageZh)

	_, _, err := parseRows(bytes.NewReader(data), "test_sheet", columns)
	requireErrorContaining(t, err, "no data found", "empty template should report no data")
}

// requireError fails the test unless err is non-nil.
func requireError(t *testing.T, err error, context string) {
	t.Helper()
	if err != nil {
		return
	}
	t.Fatalf("%s, got nil error", context)
}

// requireErrorContaining fails the test unless err contains want.
func requireErrorContaining(t *testing.T, err error, want, context string) {
	t.Helper()
	requireError(t, err, context)
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("%s, got %v", context, err)
	}
}
