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
	"strings"

	"github.com/xuri/excelize/v2"
)

func (tmpl *template[T]) ParseIndexed(r io.Reader) ([]IndexedRow[T], error) {
	dataRows, titles, err := parseRows(r, tmpl.sheet, tmpl.meta)
	if err != nil {
		return nil, err
	}

	rows := make([]IndexedRow[T], 0, len(dataRows))
	for rowIdx, cells := range dataRows {
		empty := true
		for _, cell := range cells {
			if strings.TrimSpace(cell) != "" {
				empty = false
				break
			}
		}
		if empty {
			continue
		}

		var row T
		rowErr := tmpl.parseRow(&row, cells, titles)
		rows = append(rows, IndexedRow[T]{RowIndex: rowIdx + dataStartRow, Value: row, Err: rowErr})
	}

	return rows, nil
}

// parseRows reads the data rows and validated column titles of a template file.
func parseRows(r io.Reader, sheet string, columns []Column) ([][]string, []string, error) {
	if len(columns) == 0 {
		return nil, nil, errors.New("excel columns can not be empty")
	}

	f, err := excelize.OpenReader(r, excelize.Options{RawCellValue: true})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open excel workbook: %w", err)
	}
	defer func() { _ = f.Close() }()

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get excel rows: %w", err)
	}
	if len(rows) < dataStartRow {
		return nil, nil, errors.New("no data found in excel file")
	}

	titles := rows[headerRow-1]
	if err := validateHeader(titles, rows[requiredRow-1], columns); err != nil {
		return nil, nil, err
	}

	return rows[dataStartRow-1:], titles, nil
}

// validateHeader checks each title and required label against its column in either language.
func validateHeader(titles, labels []string, columns []Column) error {
	if len(titles) != len(columns) {
		return errors.New("invalid excel template header, please download the latest template")
	}

	for idx, col := range columns {
		if titles[idx] != col.Name.in(LanguageZh) && titles[idx] != col.Name.in(LanguageEn) {
			return errors.New("invalid excel template header, please download the latest template")
		}
	}

	if len(labels) != len(columns) {
		return errors.New("outdated excel template layout, please download the latest template")
	}
	for idx, col := range columns {
		label := requiredLabel(col.Required)
		if labels[idx] != label.in(LanguageZh) && labels[idx] != label.in(LanguageEn) {
			return errors.New("outdated excel template layout, please download the latest template")
		}
	}

	return nil
}

// parseRow fills one data row, stopping at the first missing required column
// or column parse failure.
func (tmpl *template[T]) parseRow(row *T, cells, titles []string) error {
	for colIdx, col := range tmpl.columns {
		value := ""
		if colIdx < len(cells) {
			value = cells[colIdx]
		}
		if col.Required && strings.TrimSpace(value) == "" {
			return fmt.Errorf("column %q is required", titles[colIdx])
		}
		if err := col.Parse(row, value); err != nil {
			return err
		}
	}

	return nil
}
