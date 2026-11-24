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
	"unicode/utf16"

	"github.com/xuri/excelize/v2"
)

// addDropdowns adds drop-down validation to columns with options.
func addDropdowns(f *excelize.File, sheet string, columns []Column, language Language) error {
	var warningTitle, warningMessage, stopTitle, stopMessage Text
	warningTitle.Edit().Zh(dropdownWarningTitleZh).En(dropdownWarningTitleEn).Set()
	warningMessage.Edit().Zh(dropdownWarningMessageZh).En(dropdownWarningMessageEn).Set()
	stopTitle.Edit().Zh(dropdownStopTitleZh).En(dropdownStopTitleEn).Set()
	stopMessage.Edit().Zh(dropdownStopMessageZh).En(dropdownStopMessageEn).Set()

	for colIndex, col := range columns {
		if len(col.Options) == 0 {
			continue
		}

		colName, err := excelize.ColumnNumberToName(colIndex + 1)
		if err != nil {
			return fmt.Errorf("failed to resolve excel column name. col-index(%d): %w", colIndex+1, err)
		}

		dv := excelize.NewDataValidation(true)
		dv.Sqref = fmt.Sprintf("%s%d:%s%d", colName, dataStartRow, colName, excelize.TotalRows)

		if dropListUTF16Len(col.Options) <= inlineDropListLimit {
			if err := dv.SetDropList(col.Options); err != nil {
				return fmt.Errorf("failed to set excel drop list. column(%s): %w", col.Name.in(LanguageEn), err)
			}
		} else {
			sqref, err := writeOptionsSheet(f, colIndex, col.Options)
			if err != nil {
				return fmt.Errorf("failed to write excel options sheet. column(%s): %w", col.Name.in(LanguageEn), err)
			}
			dv.SetSqrefDropList(sqref)
		}

		if col.AllowCustom {
			dv.SetError(excelize.DataValidationErrorStyleWarning, warningTitle.in(language), warningMessage.in(language))
		} else {
			dv.SetError(excelize.DataValidationErrorStyleStop, stopTitle.in(language), stopMessage.in(language))
		}
		if err := f.AddDataValidation(sheet, dv); err != nil {
			return fmt.Errorf("failed to add excel data validation. column(%s): %w", col.Name.in(LanguageEn), err)
		}
	}

	return nil
}

// writeOptionsSheet stores options in a sheet and returns its range reference.
func writeOptionsSheet(f *excelize.File, colIndex int, options []string) (string, error) {
	name := fmt.Sprintf(optionsSheetNameFormat, colIndex+1)
	if _, err := f.NewSheet(name); err != nil {
		return "", fmt.Errorf("failed to create options sheet %s: %w", name, err)
	}
	for idx, option := range options {
		cell, err := excelize.CoordinatesToCellName(1, idx+1)
		if err != nil {
			return "", fmt.Errorf("failed to resolve options cell. row(%d): %w", idx+1, err)
		}
		if err := f.SetCellValue(name, cell, option); err != nil {
			return "", fmt.Errorf("failed to set options cell %s: %w", cell, err)
		}
	}
	if err := f.SetSheetVisible(name, false); err != nil {
		return "", fmt.Errorf("failed to hide options sheet %s: %w", name, err)
	}

	return fmt.Sprintf(optionsSheetRangeFormat, name, len(options)), nil
}

func dropListUTF16Len(options []string) int {
	length := 0
	for idx, option := range options {
		if idx > 0 {
			length++
		}
		for _, r := range option {
			length += utf16.RuneLen(r)
		}
	}

	return length
}
