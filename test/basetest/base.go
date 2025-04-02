/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package basetest provides a base test suite for unit tests in Go.
package basetest

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// TestSuit is a base test suite for unit tests.
type TestSuit struct {
	suite.Suite

	setupSuiteFunc    func()
	tearDownSuiteFunc func()
	cleanupFuncs      []func()
}

// SetupSuite initializes the test suite.
func (s *TestSuit) SetupSuite() {
	if s.setupSuiteFunc != nil {
		s.setupSuiteFunc()
	}
	s.T().Log("🚀 Starting test suite")
}

// TearDownSuite finalizes the test suite.
func (s *TestSuit) TearDownSuite() {
	if s.tearDownSuiteFunc != nil {
		s.tearDownSuiteFunc()
	}
	s.T().Log("✅ Finishing test suite")
}

// SetupTest initializes the test case.
func (s *TestSuit) SetupTest() {
	s.cleanupFuncs = nil
}

// TearDownTest tear down the test case.
func (s *TestSuit) TearDownTest() {
	// Execute all cleanup functions in reverse order
	for i := len(s.cleanupFuncs) - 1; i >= 0; i-- {
		s.cleanupFuncs[i]()
	}
}

// AddCleanup register a function to be executed after the test case.
func (s *TestSuit) AddCleanup(fn func()) {
	s.cleanupFuncs = append(s.cleanupFuncs, fn)
}

// AddSetupSuiteFunc register a function to be executed before the test suite.
func (s *TestSuit) AddSetupSuiteFunc(fn func()) {
	s.setupSuiteFunc = fn
}

// AddTearDownSuiteFunc register a function to be executed after the test suite.
func (s *TestSuit) AddTearDownSuiteFunc(fn func()) {
	s.tearDownSuiteFunc = fn
}

// RunTests runs the test suite.
func RunTests(t *testing.T, s suite.TestingSuite) {
	suite.Run(t, s)
}
