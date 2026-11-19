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
	"strconv"
	"testing"

	"github.com/xuri/excelize/v2"
)

type templateTestRow struct {
	IP   string
	Port int64
	Note string
}

func testTypedColumns() []TypedColumn[templateTestRow] {
	return []TypedColumn[templateTestRow]{
		{
			Column: Column{Name: "内网 IPv4", Required: true, Description: "目标主机内网 IPv4 地址。"},
			Parse:  func(row *templateTestRow, value string) error { row.IP = value; return nil },
			ValueOf: func(row *templateTestRow) any {
				return row.IP
			},
		},
		{
			Column: Column{Name: "端口", Required: true, Description: "登录端口。"},
			Parse: func(row *templateTestRow, value string) error {
				port, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return fmt.Errorf("failed to parse port. port(%s): %w", value, err)
				}
				row.Port = port

				return nil
			},
			ValueOf: func(row *templateTestRow) any {
				return row.Port
			},
		},
		{
			Column: Column{Name: "备注", Required: false, Description: "自由填写。"},
			Parse:  func(row *templateTestRow, value string) error { row.Note = value; return nil },
			ValueOf: func(row *templateTestRow) any {
				return row.Note
			},
		},
	}
}

func newTestTemplate(t *testing.T, columns []TypedColumn[templateTestRow]) *Template[templateTestRow] {
	t.Helper()
	tmpl, err := NewTemplate("typed_sheet", columns)
	if err != nil {
		t.Fatalf("newTemplate failed: %v", err)
	}

	return tmpl
}

func TestTemplateWriteParseRoundTrip(t *testing.T) {
	tmpl := newTestTemplate(t, testTypedColumns())
	samples := []templateTestRow{
		{IP: "1.1.1.1", Port: 22, Note: "first"},
		{IP: "1.1.1.2", Port: 2222, Note: "second"},
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Write(buf, samples); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	indexed, err := tmpl.ParseIndexed(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("ParseIndexed failed: %v", err)
	}
	if len(indexed) != len(samples) {
		t.Fatalf("want %d rows, got %d", len(samples), len(indexed))
	}
	for idx, item := range indexed {
		if item.Err != nil {
			t.Fatalf("row %d: unexpected parse error: %v", idx, item.Err)
		}
		if item.Value != samples[idx] {
			t.Fatalf("row %d: want %+v, got %+v", idx, samples[idx], item.Value)
		}
	}
}

func TestTemplateParseShortRowKeepsZeroValues(t *testing.T) {
	tmpl := newTestTemplate(t, testTypedColumns())
	buf := new(bytes.Buffer)
	if err := tmpl.Write(buf, []templateTestRow{{IP: "1.1.1.1", Port: 22, Note: "filled"}}); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.SetCellValue("typed_sheet", "C4", ""); err != nil {
		t.Fatal(err)
	}
	trimmed := new(bytes.Buffer)
	if err := f.Write(trimmed); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	indexed, err := tmpl.ParseIndexed(bytes.NewReader(trimmed.Bytes()))
	if err != nil {
		t.Fatalf("ParseIndexed failed: %v", err)
	}
	if len(indexed) != 1 || indexed[0].Err != nil {
		t.Fatalf("unexpected rows: %+v", indexed)
	}
	row := indexed[0].Value
	if row.IP != "1.1.1.1" || row.Port != 22 || row.Note != "" {
		t.Fatalf("unexpected row: %+v", row)
	}
}

func TestTemplateParseIndexedErrorCarriesRowIndex(t *testing.T) {
	tmpl := newTestTemplate(t, testTypedColumns())
	buf := new(bytes.Buffer)
	samples := []templateTestRow{
		{IP: "1.1.1.1", Port: 22},
		{IP: "1.1.1.2", Port: 22},
	}
	if err := tmpl.Write(buf, samples); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.SetCellValue("typed_sheet", "B5", "not-a-port"); err != nil {
		t.Fatal(err)
	}
	broken := new(bytes.Buffer)
	if err := f.Write(broken); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	indexed, err := tmpl.ParseIndexed(bytes.NewReader(broken.Bytes()))
	if err != nil {
		t.Fatalf("ParseIndexed failed: %v", err)
	}
	if len(indexed) != 2 {
		t.Fatalf("want 2 rows, got %d", len(indexed))
	}
	if indexed[0].Err != nil {
		t.Fatalf("first row should parse, got %v", indexed[0].Err)
	}
	if indexed[1].RowIndex != dataStartRow+1 {
		t.Fatalf("error row should carry index %d, got %d", dataStartRow+1, indexed[1].RowIndex)
	}
	requireErrorContaining(t, indexed[1].Err, "failed to parse port", "row error should carry the column parse failure")
}

func TestTemplateRejectsNilCallbacks(t *testing.T) {
	columns := testTypedColumns()
	columns[0].ValueOf = nil
	_, err := NewTemplate("typed_sheet", columns)
	requireError(t, err, "nil ValueOf should be rejected on construction")

	columns = testTypedColumns()
	columns[0].Parse = nil
	_, err = NewTemplate("typed_sheet", columns)
	requireError(t, err, "nil Parse should be rejected on construction")
}

func TestTemplateRejectsEmptyColumns(t *testing.T) {
	_, err := NewTemplate[templateTestRow]("typed_sheet", nil)
	requireError(t, err, "empty columns should be rejected on construction")
}

func TestTemplateCopiesColumns(t *testing.T) {
	columns := testTypedColumns()
	tmpl := newTestTemplate(t, columns)
	columns[0].Column.Name = "被修改的表头"

	buf := new(bytes.Buffer)
	if err := tmpl.Write(buf, []templateTestRow{{IP: "1.1.1.1", Port: 22}}); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	value, err := f.GetCellValue("typed_sheet", "A1")
	if err != nil {
		t.Fatal(err)
	}
	if value != "内网 IPv4" {
		t.Fatalf("template should keep its own columns, got %q", value)
	}
}

func TestTemplateSkipsEmptyRows(t *testing.T) {
	tmpl := newTestTemplate(t, testTypedColumns())
	buf := new(bytes.Buffer)
	samples := []templateTestRow{
		{IP: "1.1.1.1", Port: 22},
		{IP: "1.1.1.2", Port: 22},
		{IP: "1.1.1.3", Port: 22},
	}
	if err := tmpl.Write(buf, samples); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	for _, cell := range []string{"A5", "B5", "C5"} {
		if err := f.SetCellValue("typed_sheet", cell, ""); err != nil {
			t.Fatal(err)
		}
	}
	withBlank := new(bytes.Buffer)
	if err := f.Write(withBlank); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	indexed, err := tmpl.ParseIndexed(bytes.NewReader(withBlank.Bytes()))
	if err != nil {
		t.Fatalf("ParseIndexed failed: %v", err)
	}
	if len(indexed) != 2 || indexed[0].RowIndex != 4 || indexed[1].RowIndex != 6 {
		t.Fatalf("blank row should be skipped with original row numbers, got %+v", indexed)
	}
}
