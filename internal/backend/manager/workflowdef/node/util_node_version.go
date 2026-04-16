/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
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

func autoSelectVersion(nCtx contextx.IContext, versionParam CheckAndSelectVersionParam) (string, error) {
	plat, err := platfmt.Normalize(string(versionParam.OSType), string(versionParam.CPUArch))
	if err != nil {
		return "", fmt.Errorf("invalid platform: %w", err)
	}

	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Platform:   []platfmt.Platform{plat},
			Generation: []types.Generation{versionParam.Generation},
			AsDefault:  []bool{true},
			Enabled:    []bool{true},
		},
	}

	switch versionParam.ReleaseType {
	case types.ReleaseTypeAgent:
		r, _, err := versionParam.daoRelease.ListReleaseAgent(nCtx, types.UnlimitedPage(), cond)
		if err != nil {
			return "", fmt.Errorf("failed to list default agent releases: %w", err)
		}

		if len(r) == 0 {
			return "", fmt.Errorf("failed to list default agent releases for platform. no default agent release found. condition(%+v)", cond)
		}

		if len(r) > 1 {
			return "", fmt.Errorf(
				"failed to list default agent releases for platform. multiple default agent releases found. condition(%+v)", cond)
		}

		return r[0].Version, nil

	case types.ReleaseTypeProxy:
		r, _, err := versionParam.daoRelease.ListReleaseProxy(nCtx, types.UnlimitedPage(), cond)
		if err != nil {
			return "", fmt.Errorf("failed to list default proxy releases: %w", err)
		}

		if len(r) == 0 {
			return "", fmt.Errorf("failed to list default proxy releases for platform. no default proxy release found. condition(%+v)", cond)
		}

		if len(r) > 1 {
			return "", fmt.Errorf(
				"failed to list default proxy releases for platform. multiple default proxy releases found. condition(%+v)", cond)
		}

		return r[0].Version, nil

	default:
		return "", fmt.Errorf("invalid release type: %v", versionParam.ReleaseType)
	}
}

func checkVersionAvailability(nCtx contextx.IContext, versionParam CheckAndSelectVersionParam) error {
	plat, err := platfmt.Normalize(string(versionParam.OSType), string(versionParam.CPUArch))
	if err != nil {
		return fmt.Errorf("invalid platform: %w", err)
	}

	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Platform:   []platfmt.Platform{plat},
			Generation: []types.Generation{versionParam.Generation},
			Version:    []string{versionParam.Version},
			Enabled:    []bool{true},
		},
	}

	switch versionParam.ReleaseType {
	case types.ReleaseTypeAgent:
		r, _, err := versionParam.daoRelease.ListReleaseAgent(nCtx, types.UnlimitedPage(), cond)
		if err != nil {
			return fmt.Errorf("failed to list default agent releases: %w", err)
		}

		if len(r) == 0 {
			return fmt.Errorf("failed to list default agent releases for version. no agent release found. condition(%+v)", cond)
		}

		if len(r) > 1 {
			return fmt.Errorf(
				"failed to list default agent releases for version. multiple default agent releases found. condition(%+v)", cond)
		}

		return nil

	case types.ReleaseTypeProxy:
		r, _, err := versionParam.daoRelease.ListReleaseProxy(nCtx, types.UnlimitedPage(), cond)
		if err != nil {
			return fmt.Errorf("failed to list default proxy releases: %w", err)
		}

		if len(r) == 0 {
			return fmt.Errorf("failed to list default proxy releases for version. no proxy release found. condition(%+v)", cond)
		}

		if len(r) > 1 {
			return fmt.Errorf(
				"failed to list default proxy releases for version. multiple default proxy releases found. condition(%+v)", cond)
		}

		return nil

	default:
		return fmt.Errorf("invalid release type: %v", versionParam.ReleaseType)
	}
}
