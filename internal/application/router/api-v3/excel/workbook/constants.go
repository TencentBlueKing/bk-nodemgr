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

const (
	// Template layout rows.
	headerRow    = 1
	requiredRow  = 2
	describeRow  = 3
	dataStartRow = 4

	// Default sheet name.
	defaultSheetName = "Sheet1"
)

const (
	// Column width and row height.
	minColWidth    = 24
	maxColWidth    = 50
	cellPadding    = 2
	lineHeight     = 15
	rowPadding     = 6
	minUsableWidth = 1
	nonASCIIWidth  = 2

	// Font sizes.
	headerFontSize = 14
	bodyFontSize   = 11

	// Fill pattern.
	fillTypePattern  = "pattern"
	fillPatternSolid = 1

	// Alignment.
	alignCenter = "center"
	alignLeft   = "left"
	alignTop    = "top"

	// Colors.
	colorBlue      = "538DD5"
	colorLightGray = "D9D9D9"
	colorDarkRed   = "C0504D"
	colorOrange    = "E26B0A"
)

const (
	// Inline list formula limit in UTF-16 units.
	inlineDropListLimit = 250

	// Options sheet.
	optionsSheetNameFormat  = "_options_%d"
	optionsSheetRangeFormat = "'%s'!$A$1:$A$%d"
)

const (
	// Template copy rendered into templates (zh-CN).
	labelRequired = "必填"
	labelOptional = "可选"

	dropdownWarningTitle   = "输入提示"
	dropdownWarningMessage = "建议从下拉列表中选择值，也可填写自定义值"
	dropdownStopTitle      = "输入不合法"
	dropdownStopMessage    = "请从下拉列表中选择值"
)
