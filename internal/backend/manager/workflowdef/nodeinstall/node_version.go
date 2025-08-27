/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstall

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// CheckAndSelectVersionParam defines the version param.
type CheckAndSelectVersionParam struct {
	daoRelease  release.IStorage
	ReleaseType types.ReleaseType
	Generation  types.Generation
	OSType      criteria.OSType
	CPUArch     criteria.CPUArch
	Version     string
}

func autoSelectVersion(ctx context.Context, versionParam CheckAndSelectVersionParam) (string, error) {
	plat := platform.NewPlatform(versionParam.OSType, versionParam.CPUArch)

	releaseType := versionParam.ReleaseType
	gen := versionParam.Generation

	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Platform:   []platform.Platform{plat},
			Generation: []types.Generation{gen},
			AsDefault:  []bool{true},
			Enabled:    []bool{true},
		},
	}

	releases, num, err := versionParam.daoRelease.ListRelease(ctx, releaseType, types.UnlimitedPage(), cond)
	if err != nil {
		return "", fmt.Errorf("failed to list default releases, err: %w", err)
	}
	if num == 0 {
		return "", fmt.Errorf("failed to list default releases for platform. no default release found. platform(%v)", plat)
	}

	if num > 1 {
		return "", fmt.Errorf(
			"failed to list default releases for platform. multiple default releases found. platform(%v)", plat)
	}

	return releases[0].Version, nil
}

func checkVersionAvailability(ctx context.Context, versionParam CheckAndSelectVersionParam) error {
	plat, err := platform.Normalize(string(versionParam.OSType), string(versionParam.CPUArch))
	if err != nil {
		return fmt.Errorf("invalid platform, err: %w", err)
	}

	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Platform:   []platform.Platform{plat},
			Generation: []types.Generation{versionParam.Generation},
			Version:    []string{versionParam.Version},
			Enabled:    []bool{true},
		},
	}
	num, err := versionParam.daoRelease.CountRelease(ctx, versionParam.ReleaseType, cond)
	if err != nil {
		return fmt.Errorf("failed to check release version,err: %w", err)
	}

	if num == 0 {
		return fmt.Errorf(
			"failed to check release version. no release found. version(%v)", versionParam.Version)
	}

	if num > 1 {
		return fmt.Errorf(
			"failed to check release version. multiple releases found. version(%v)", versionParam.Version)
	}

	return nil
}
