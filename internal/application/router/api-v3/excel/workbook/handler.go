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
)

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

// IHandler renders and parses a template with typed rows.
type IHandler[T any] interface {
	// Write renders the template with the given data rows.
	Write(out io.Writer, rows []T) error
	// ParseIndexed reads typed data rows with their Excel row numbers and row-level errors.
	ParseIndexed(r io.Reader) ([]IndexedRow[T], error)
}

type template[T any] struct {
	sheet   string
	columns []TypedColumn[T]
	meta    []Column
}

// NewTemplate creates a handler for the given sheet and typed columns.
func NewTemplate[T any](sheet string, columns []TypedColumn[T]) (IHandler[T], error) {
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

	return &template[T]{sheet: sheet, columns: copied, meta: meta}, nil
}
