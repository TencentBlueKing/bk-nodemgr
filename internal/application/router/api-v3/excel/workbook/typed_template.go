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
)

// TypedColumn binds one template column to row type T.
type TypedColumn[T any] struct {
	Column
	// Parse assigns a raw cell value into the row.
	Parse func(row *T, value string) error
	// ValueOf returns the cell value to render.
	ValueOf func(row *T) any
}

// IndexedRow pairs a parsed row with its Excel row number.
type IndexedRow[T any] struct {
	// RowIndex is the Excel row number.
	RowIndex int
	// Value is the parsed row.
	Value T
	// Err records the row-level parse error; nil means the row parsed successfully.
	Err error
}

// Template renders and parses a sheet with typed columns.
type Template[T any] struct {
	sheet   string
	columns []TypedColumn[T]
	meta    []Column
}

// NewTemplate creates a typed template.
func NewTemplate[T any](sheet string, columns []TypedColumn[T]) (*Template[T], error) {
	if len(columns) == 0 {
		return nil, errors.New("excel columns can not be empty")
	}
	for _, col := range columns {
		if col.Parse == nil {
			return nil, fmt.Errorf("excel column %q has no Parse", col.Name)
		}
		if col.ValueOf == nil {
			return nil, fmt.Errorf("excel column %q has no ValueOf", col.Name)
		}
	}

	copied := make([]TypedColumn[T], len(columns))
	copy(copied, columns)
	meta := make([]Column, len(columns))
	for idx, col := range copied {
		meta[idx] = col.Column
	}

	return &Template[T]{sheet: sheet, columns: copied, meta: meta}, nil
}

// Write renders the template with the given data rows.
func (tmpl *Template[T]) Write(out io.Writer, rows []T) error {
	cells := make([][]any, len(rows))
	for rowIdx := range rows {
		row := make([]any, len(tmpl.columns))
		for colIdx, col := range tmpl.columns {
			row[colIdx] = col.ValueOf(&rows[rowIdx])
		}
		cells[rowIdx] = row
	}

	return WriteTemplate(out, tmpl.sheet, tmpl.meta, cells)
}

// ParseIndexed reads typed data rows with their Excel row numbers. Row-level
// parse errors are stored on IndexedRow.Err instead of failing the whole read.
func (tmpl *Template[T]) ParseIndexed(r io.Reader) ([]IndexedRow[T], error) {
	dataRows, err := ParseRows(r, tmpl.sheet, tmpl.meta)
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
		rowErr := tmpl.parseRow(&row, cells)
		rows = append(rows, IndexedRow[T]{RowIndex: rowIdx + dataStartRow, Value: row, Err: rowErr})
	}

	return rows, nil
}

// parseRow fills one data row, stopping at the first missing required column
// or column parse failure.
func (tmpl *Template[T]) parseRow(row *T, cells []string) error {
	for colIdx, col := range tmpl.columns {
		value := ""
		if colIdx < len(cells) {
			value = cells[colIdx]
		}
		if col.Required && strings.TrimSpace(value) == "" {
			return fmt.Errorf("column %q is required", col.Name)
		}
		if err := col.Parse(row, value); err != nil {
			return err
		}
	}

	return nil
}
