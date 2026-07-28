/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logger

import (
	"fmt"
	"strings"
	"testing"
)

type capturePrinter struct {
	message string
}

func (p *capturePrinter) Log(_ Category, _ Level, _ int, format string, args ...interface{}) {
	p.message = fmt.Sprintf(format, args...)
}

func TestWithFieldPercentSignsRemainLiteral(t *testing.T) {
	oldPrinter := printer
	capture := &capturePrinter{}
	printer = capture
	defer func() {
		printer = oldPrinter
	}()

	var assigned string
	fieldValue := "http://example.com/login?bk_token=bkcrypt%gxxx%D"
	Logger{Depth: 1}.Sys().With("url", fieldValue).AssignWhenLogging(&assigned).Info("get response data")

	for _, message := range []string{capture.message, assigned} {
		if !strings.Contains(message, fieldValue) {
			t.Fatalf("expected message to contain literal field value %q, got %q", fieldValue, message)
		}
		if strings.Contains(message, "MISSING") || strings.Contains(message, "%!") {
			t.Fatalf("expected message not to contain fmt missing marker, got %q", message)
		}
	}
}
