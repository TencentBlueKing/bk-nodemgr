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
	"errors"
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

func (tmpl *template[T]) Write(out io.Writer, rows []T) error {
	cells := make([][]any, len(rows))
	for rowIdx := range rows {
		row := make([]any, len(tmpl.columns))
		for colIdx, col := range tmpl.columns {
			row[colIdx] = col.ValueOf(&rows[rowIdx])
		}
		cells[rowIdx] = row
	}

	return writeTemplate(out, tmpl.sheet, tmpl.meta, cells, tmpl.language)
}

// writeTemplate renders the template to out; data rows align with columns by index.
func writeTemplate(out io.Writer, sheet string, columns []Column, rows [][]any, language Language) error {
	if len(columns) == 0 {
		return errors.New("excel columns can not be empty")
	}

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	if err := f.SetSheetName(defaultSheetName, sheet); err != nil {
		return fmt.Errorf("failed to rename default excel sheet: %w", err)
	}

	if err := writeHeaderRows(f, sheet, columns, language); err != nil {
		return err
	}
	if err := applyLayout(f, sheet, columns, language); err != nil {
		return err
	}
	if err := addDropdowns(f, sheet, columns, language); err != nil {
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

func writeHeaderRows(f *excelize.File, sheet string, columns []Column, language Language) error {
	for colIndex, col := range columns {
		label := requiredLabel(col.Required)
		values := map[int]string{
			headerRow:   col.Name.in(language),
			requiredRow: label.in(language),
			describeRow: col.Description.in(language),
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
