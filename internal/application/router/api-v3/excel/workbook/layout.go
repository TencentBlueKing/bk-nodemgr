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
	"fmt"
	"unicode"

	"github.com/xuri/excelize/v2"
)

func applyLayout(f *excelize.File, sheet string, columns []Column) error {
	widths := make([]float64, len(columns))
	for idx, col := range columns {
		needed := max(textDisplayWidth(col.Name), textDisplayWidth(requiredLabel(col.Required)), textDisplayWidth(col.Description))
		widths[idx] = float64(min(max(needed+cellPadding, minColWidth), maxColWidth))
	}

	lastCol, err := excelize.ColumnNumberToName(len(columns))
	if err != nil {
		return fmt.Errorf("failed to resolve last excel column: %w", err)
	}

	if err := applyHeaderStyles(f, sheet, columns, lastCol); err != nil {
		return err
	}

	for colIndex := range columns {
		colName, err := excelize.ColumnNumberToName(colIndex + 1)
		if err != nil {
			return fmt.Errorf("failed to resolve excel column name. col-index(%d): %w", colIndex+1, err)
		}
		if err := f.SetColWidth(sheet, colName, colName, widths[colIndex]); err != nil {
			return fmt.Errorf("failed to set excel column width. column(%s): %w", colName, err)
		}
	}

	headerValues := make([]string, len(columns))
	requiredValues := make([]string, len(columns))
	describeValues := make([]string, len(columns))
	for idx, col := range columns {
		headerValues[idx] = col.Name
		requiredValues[idx] = requiredLabel(col.Required)
		describeValues[idx] = col.Description
	}
	for row, values := range map[int][]string{
		headerRow:   headerValues,
		requiredRow: requiredValues,
		describeRow: describeValues,
	} {
		lines := 1
		for idx, value := range values {
			usableWidth := int(widths[idx]) - cellPadding
			if usableWidth < minUsableWidth {
				usableWidth = minUsableWidth
			}
			lineCount := (textDisplayWidth(value) + usableWidth - 1) / usableWidth
			lines = max(lines, lineCount)
		}
		if err := f.SetRowHeight(sheet, row, float64(lines*lineHeight+rowPadding)); err != nil {
			return fmt.Errorf("failed to set excel row height. row(%d): %w", row, err)
		}
	}

	return nil
}

// nolint: mnd
func applyHeaderStyles(f *excelize.File, sheet string, columns []Column, lastCol string) error {
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: headerFontSize, Color: colorBlue},
		Fill:      excelize.Fill{Type: fillTypePattern, Pattern: fillPatternSolid, Color: []string{colorLightGray}},
		Alignment: &excelize.Alignment{Horizontal: alignCenter, Vertical: alignCenter, WrapText: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create excel header style: %w", err)
	}
	requiredStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: bodyFontSize, Color: colorDarkRed},
		Fill:      excelize.Fill{Type: fillTypePattern, Pattern: fillPatternSolid, Color: []string{colorLightGray}},
		Alignment: &excelize.Alignment{Horizontal: alignCenter, Vertical: alignCenter, WrapText: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create excel required style: %w", err)
	}
	optionalStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: bodyFontSize, Color: colorOrange},
		Fill:      excelize.Fill{Type: fillTypePattern, Pattern: fillPatternSolid, Color: []string{colorLightGray}},
		Alignment: &excelize.Alignment{Horizontal: alignCenter, Vertical: alignCenter, WrapText: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create excel optional style: %w", err)
	}
	describeStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: bodyFontSize},
		Fill:      excelize.Fill{Type: fillTypePattern, Pattern: fillPatternSolid, Color: []string{colorLightGray}},
		Alignment: &excelize.Alignment{Horizontal: alignLeft, Vertical: alignTop, WrapText: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create excel describe style: %w", err)
	}

	if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", headerRow), fmt.Sprintf("%s%d", lastCol, headerRow), headerStyle); err != nil {
		return fmt.Errorf("failed to apply excel header style: %w", err)
	}
	for colIndex, col := range columns {
		styleID := requiredStyle
		if !col.Required {
			styleID = optionalStyle
		}
		cell, _ := excelize.CoordinatesToCellName(colIndex+1, requiredRow)
		if err := f.SetCellStyle(sheet, cell, cell, styleID); err != nil {
			return fmt.Errorf("failed to apply excel required style on %s: %w", cell, err)
		}
	}
	if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", describeRow), fmt.Sprintf("%s%d", lastCol, describeRow), describeStyle); err != nil {
		return fmt.Errorf("failed to apply excel describe style: %w", err)
	}

	return nil
}

// textDisplayWidth returns the display width of value.
func textDisplayWidth(value string) int {
	width := 0
	for _, r := range value {
		if r > unicode.MaxASCII {
			width += nonASCIIWidth
			continue
		}
		width++
	}

	return width
}
