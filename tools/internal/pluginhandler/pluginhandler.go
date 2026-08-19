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

// Package pluginhandler provides the plugin handler interface and implementation.
package pluginhandler

import (
	"context"
)

// IPluginHandler plugin handler interface.
type IPluginHandler interface {
	// Group return plugin group.
	Group() string

	// FS return plugin file system handler.
	FS() IPluginFSHandler

	// Process return plugin process handler.
	Process() IPluginProcessHandler
}

// IPluginFSHandler plugin file system handler interface.
type IPluginFSHandler interface {
	// CheckIntegrity check the integrity of plugin file system.
	CheckIntegrity(ctx context.Context) error

	// Init init plugin file system.
	Init() error

	// Purge purge plugin file system.
	Purge(ctx context.Context) error

	// CopyConfigDir copy config files to plugin file system.
	CopyConfigDir(ctx context.Context, configFileAbsDir string) error

	// UnpackReleasePackage unpack release package to plugin file system.
	// if keepOldFileAsTmp is true, will move old-existing file to a tmp file in same directory,
	// to keep the process running. Call Clean() to clean all the tmp files.
	UnpackReleasePackage(ctx context.Context, releasePkgAbsPath string, keepOldFileAsTmp bool) error

	// Clean cleans all the tmp files and tools in plugin file system.
	Clean(ctx context.Context) error
}

// IPluginProcessHandler plugin process handler interface.
type IPluginProcessHandler interface {
	// RunCmd run command in plugin process handler.
	RunCmd(ctx context.Context, cmd string, pidDir string) error

	// Stop stop plugin process handler.
	Stop(ctx context.Context, pidDir string) error
}
