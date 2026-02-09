/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package action

import (
	"sync"
	"testing"
)

func TestLogBuilder_ZhAndEn(t *testing.T) {
	data := &InstanceData{}

	// Test: Both Zh and En are set
	data.Log().Zh("中文消息").En("English message").Info()

	if len(data.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(data.Messages))
	}

	msg := data.Messages[0]
	if msg.TextZh != "中文消息" {
		t.Errorf("expected TextZh='中文消息', got '%s'", msg.TextZh)
	}
	if msg.TextEn != "English message" {
		t.Errorf("expected TextEn='English message', got '%s'", msg.TextEn)
	}
	if msg.Level != "INFO" {
		t.Errorf("expected Level='INFO', got '%s'", msg.Level)
	}
}

func TestLogBuilder_OnlyZh(t *testing.T) {
	data := &InstanceData{}

	// Test: Only Zh is set, En should be auto-filled with Zh
	data.Log().Zh("只有中文").Info()

	if len(data.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(data.Messages))
	}

	msg := data.Messages[0]
	if msg.TextZh != "只有中文" {
		t.Errorf("expected TextZh='只有中文', got '%s'", msg.TextZh)
	}
	if msg.TextEn != "只有中文" {
		t.Errorf("expected TextEn='只有中文' (auto-filled), got '%s'", msg.TextEn)
	}
}

func TestLogBuilder_OnlyEn(t *testing.T) {
	data := &InstanceData{}

	// Test: Only En is set, Zh should be auto-filled with En
	data.Log().En("Only English").Warn()

	if len(data.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(data.Messages))
	}

	msg := data.Messages[0]
	if msg.TextZh != "Only English" {
		t.Errorf("expected TextZh='Only English' (auto-filled), got '%s'", msg.TextZh)
	}
	if msg.TextEn != "Only English" {
		t.Errorf("expected TextEn='Only English', got '%s'", msg.TextEn)
	}
	if msg.Level != "WARN" {
		t.Errorf("expected Level='WARN', got '%s'", msg.Level)
	}
}

func TestLogBuilder_NeitherSet(t *testing.T) {
	data := &InstanceData{}

	// Test: Neither Zh nor En is set, should use level as default
	data.Log().Error()

	if len(data.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(data.Messages))
	}

	msg := data.Messages[0]
	if msg.TextZh != "ERROR" {
		t.Errorf("expected TextZh='ERROR' (default), got '%s'", msg.TextZh)
	}
	if msg.TextEn != "ERROR" {
		t.Errorf("expected TextEn='ERROR' (default), got '%s'", msg.TextEn)
	}
	if msg.Level != "ERROR" {
		t.Errorf("expected Level='ERROR', got '%s'", msg.Level)
	}
}

func TestLogBuilder_WithFormatArgs(t *testing.T) {
	data := &InstanceData{}

	// Test: Format string with arguments
	data.Log().
		Zh("安装节点，agent-id(%s)", "agent-123").
		En("installing node, agent-id(%s)", "agent-123").
		Info()

	if len(data.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(data.Messages))
	}

	msg := data.Messages[0]
	if msg.TextZh != "安装节点，agent-id(agent-123)" {
		t.Errorf("expected TextZh='安装节点，agent-id(agent-123)', got '%s'", msg.TextZh)
	}
	if msg.TextEn != "installing node, agent-id(agent-123)" {
		t.Errorf("expected TextEn='installing node, agent-id(agent-123)', got '%s'", msg.TextEn)
	}
}

func TestLogBuilder_EmptyString(t *testing.T) {
	data := &InstanceData{}

	// Test: Empty string is treated as "set" (zhSet=true)
	data.Log().Zh("").Info()

	if len(data.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(data.Messages))
	}

	msg := data.Messages[0]
	if msg.TextZh != "" {
		t.Errorf("expected TextZh='', got '%s'", msg.TextZh)
	}
	if msg.TextEn != "" {
		t.Errorf("expected TextEn='' (auto-filled from Zh), got '%s'", msg.TextEn)
	}
}

func TestLogBuilder_ChainOrder(t *testing.T) {
	data := &InstanceData{}

	// Test: En().Zh() order should work the same as Zh().En()
	data.Log().En("English first").Zh("中文后").Info()

	if len(data.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(data.Messages))
	}

	msg := data.Messages[0]
	if msg.TextZh != "中文后" {
		t.Errorf("expected TextZh='中文后', got '%s'", msg.TextZh)
	}
	if msg.TextEn != "English first" {
		t.Errorf("expected TextEn='English first', got '%s'", msg.TextEn)
	}
}

func TestLogBuilder_Concurrent(t *testing.T) {
	data := &InstanceData{}
	const numGoroutines = 100

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			data.Log().Zh("消息 %d", n).En("message %d", n).Info()
		}(i)
	}
	wg.Wait()

	if len(data.Messages) != numGoroutines {
		t.Errorf("expected %d messages, got %d", numGoroutines, len(data.Messages))
	}
}
