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

import "fmt"

// Language selects the language used to write a template.
type Language string

const (
	// LanguageZh selects Simplified Chinese.
	LanguageZh Language = "zh"
	// LanguageEn selects English.
	LanguageEn Language = "en"
)

// Text holds the Simplified Chinese and English versions of workbook copy.
type Text struct {
	zh string
	en string
}

// ITextHandler edits bilingual copy for a Text field.
type ITextHandler interface {
	// Zh sets the Simplified Chinese text with optional format arguments.
	Zh(format string, args ...any) ITextHandler
	// En sets the English text with optional format arguments.
	En(format string, args ...any) ITextHandler
	// Set writes the edited copy to the field.
	Set()
}

type textHandler struct {
	target *Text
	text   Text
}

// Edit returns a handler that writes to text when Set is called.
func (text *Text) Edit() ITextHandler {
	return &textHandler{target: text, text: *text}
}

func (handler *textHandler) Zh(format string, args ...any) ITextHandler {
	handler.text.zh = formatText(format, args)
	return handler
}

func (handler *textHandler) En(format string, args ...any) ITextHandler {
	handler.text.en = formatText(format, args)
	return handler
}

func (handler *textHandler) Set() {
	*handler.target = handler.text
}

func (text *Text) in(language Language) string {
	if language == LanguageEn {
		return text.en
	}

	return text.zh
}

func formatText(format string, args []any) string {
	if len(args) == 0 {
		return format
	}

	return fmt.Sprintf(format, args...)
}
