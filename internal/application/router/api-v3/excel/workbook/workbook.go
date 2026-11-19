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

// Package workbook renders and parses column-oriented Excel templates with a
// three-row header and drop-down validations; rendered copy is zh-CN.
package workbook

import (
	"errors"
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

func requiredLabel(required bool) string {
	if required {
		return labelRequired
	}

	return labelOptional
}

// Column describes one template column.
type Column struct {
	// Name is the header title (row 1).
	Name string
	// Required marks the column as required.
	Required bool
	// Description is the filling guidance.
	Description string
	// Options lists drop-down choices; empty means free text.
	Options []string
	// AllowCustom allows values outside Options.
	AllowCustom bool
}

// WriteTemplate renders the template to out; data rows align with columns by index.
func WriteTemplate(out io.Writer, sheet string, columns []Column, rows [][]any) error {
	if len(columns) == 0 {
		return errors.New("excel columns can not be empty")
	}

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	if err := f.SetSheetName(defaultSheetName, sheet); err != nil {
		return fmt.Errorf("failed to rename default excel sheet: %w", err)
	}

	if err := writeHeaderRows(f, sheet, columns); err != nil {
		return err
	}
	if err := applyLayout(f, sheet, columns); err != nil {
		return err
	}
	if err := addDropdowns(f, sheet, columns); err != nil {
		return err
	}

	if err := writeDataRows(f, sheet, columns, rows); err != nil {
		return err
	}

	if err := f.Write(out); err != nil {
		return fmt.Errorf("failed to write excel workbook: %w", err)
	}

	return nil
}

// ParseRows reads the data rows of a template file.
func ParseRows(r io.Reader, sheet string, columns []Column) ([][]string, error) {
	if len(columns) == 0 {
		return nil, errors.New("excel columns can not be empty")
	}

	f, err := excelize.OpenReader(r, excelize.Options{RawCellValue: true})
	if err != nil {
		return nil, fmt.Errorf("failed to open excel workbook: %w", err)
	}
	defer func() { _ = f.Close() }()

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("failed to get excel rows: %w", err)
	}
	if len(rows) < dataStartRow {
		return nil, errors.New("no data found in excel file")
	}

	if len(rows[0]) != len(columns) {
		return nil, fmt.Errorf("invalid excel template header. want(%d) got(%d)", len(columns), len(rows[0]))
	}
	for idx, col := range columns {
		if rows[0][idx] != col.Name {
			return nil, fmt.Errorf("invalid excel template header. col(%d) want(%q) got(%q)", idx+1, col.Name, rows[0][idx])
		}
	}

	if len(rows[1]) != len(columns) {
		return nil, errors.New("outdated excel template layout, please download the latest template")
	}
	for idx, col := range columns {
		if rows[1][idx] != requiredLabel(col.Required) {
			return nil, errors.New("outdated excel template layout, please download the latest template")
		}
	}

	return rows[dataStartRow-1:], nil
}

func writeHeaderRows(f *excelize.File, sheet string, columns []Column) error {
	for colIndex, col := range columns {
		values := map[int]string{
			headerRow:   col.Name,
			requiredRow: requiredLabel(col.Required),
			describeRow: col.Description,
		}
		for row, value := range values {
			cell, err := excelize.CoordinatesToCellName(colIndex+1, row)
			if err != nil {
				return fmt.Errorf("failed to resolve excel cell. col-index(%d) row(%d): %w", colIndex+1, row, err)
			}
			if err := f.SetCellValue(sheet, cell, value); err != nil {
				return fmt.Errorf("failed to set excel cell %s: %w", cell, err)
			}
		}
	}

	return nil
}

func writeDataRows(f *excelize.File, sheet string, columns []Column, rows [][]any) error {
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			if colIndex >= len(columns) {
				break
			}
			cell, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+dataStartRow)
			if err != nil {
				return fmt.Errorf("failed to resolve excel data cell. row(%d) col(%d): %w", rowIndex, colIndex, err)
			}
			if err := f.SetCellValue(sheet, cell, value); err != nil {
				return fmt.Errorf("failed to set excel data cell %s: %w", cell, err)
			}
		}
	}

	return nil
}
