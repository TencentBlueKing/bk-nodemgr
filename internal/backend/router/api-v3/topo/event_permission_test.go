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

package topo

import (
	"errors"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestNarrowTopoEventCondition_NarrowsAuthorizedNetworkUnits(t *testing.T) {
	t.Helper()

	calledNetworkAreaNarrower := false
	calledAccessPointNarrower := false
	condition, err := narrowTopoEventCondition(
		nil,
		func(requestedIDs []int64) ([]int64, bool, error) {
			calledNetworkAreaNarrower = true
			if len(requestedIDs) != 0 {
				t.Fatalf("expected no requested network area ids, got %v", requestedIDs)
			}

			return nil, true, nil
		},
		func(requestedIDs []int64) ([]int64, bool, error) {
			if len(requestedIDs) != 0 {
				t.Fatalf("expected no requested network unit ids, got %v", requestedIDs)
			}

			return []int64{11, 12}, false, nil
		},
		func(requestedIDs []int64) ([]int64, bool, error) {
			calledAccessPointNarrower = true
			return requestedIDs, false, nil
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !calledNetworkAreaNarrower {
		t.Fatal("expected network area narrowing to run")
	}
	if calledAccessPointNarrower {
		t.Fatal("expected access point narrowing to be skipped when no access point ids are requested")
	}
	if condition == nil || condition.ExactInclude == nil {
		t.Fatal("expected narrowed condition to initialize exact include fields")
	}
	if len(condition.ExactInclude.NetworkUnitID) != 2 || condition.ExactInclude.NetworkUnitID[0] != 11 || condition.ExactInclude.NetworkUnitID[1] != 12 {
		t.Fatalf("unexpected narrowed network unit ids: %v", condition.ExactInclude.NetworkUnitID)
	}
}

func TestNarrowTopoEventCondition_NarrowsAccessPointsAlongsideNetworkUnits(t *testing.T) {
	t.Helper()

	calledNetworkUnitNarrower := false
	condition, err := narrowTopoEventCondition(
		&types.TopoEventCondition{
			ExactInclude: &types.TopoEventExactFields{
				NetworkAreaID: []int64{3, 4},
				AccessPointID: []int64{101, 102},
			},
		},
		func(requestedIDs []int64) ([]int64, bool, error) {
			if len(requestedIDs) != 2 || requestedIDs[0] != 3 || requestedIDs[1] != 4 {
				t.Fatalf("unexpected requested network area ids: %v", requestedIDs)
			}

			return []int64{4}, false, nil
		},
		func(requestedIDs []int64) ([]int64, bool, error) {
			calledNetworkUnitNarrower = true
			if len(requestedIDs) != 0 {
				t.Fatalf("expected no requested network unit ids, got %v", requestedIDs)
			}

			return []int64{7}, false, nil
		},
		func(requestedIDs []int64) ([]int64, bool, error) {
			if len(requestedIDs) != 2 || requestedIDs[0] != 101 || requestedIDs[1] != 102 {
				t.Fatalf("unexpected requested access point ids: %v", requestedIDs)
			}

			return []int64{101}, false, nil
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !calledNetworkUnitNarrower {
		t.Fatal("expected network unit narrowing to run before access point narrowing")
	}
	if condition == nil || condition.ExactInclude == nil {
		t.Fatal("expected exact include fields after narrowing")
	}
	if len(condition.ExactInclude.NetworkAreaID) != 1 || condition.ExactInclude.NetworkAreaID[0] != 4 {
		t.Fatalf("unexpected narrowed network area ids: %v", condition.ExactInclude.NetworkAreaID)
	}
	if len(condition.ExactInclude.NetworkUnitID) != 1 || condition.ExactInclude.NetworkUnitID[0] != 7 {
		t.Fatalf("unexpected narrowed network unit ids: %v", condition.ExactInclude.NetworkUnitID)
	}
	if len(condition.ExactInclude.AccessPointID) != 1 || condition.ExactInclude.AccessPointID[0] != 101 {
		t.Fatalf("unexpected narrowed access point ids: %v", condition.ExactInclude.AccessPointID)
	}
}

func TestNarrowTopoEventCondition_PropagatesNetworkUnitError(t *testing.T) {
	t.Helper()

	expectedErr := errors.New("permission denied")
	_, err := narrowTopoEventCondition(
		&types.TopoEventCondition{},
		func(requestedIDs []int64) ([]int64, bool, error) {
			return nil, true, nil
		},
		func(requestedIDs []int64) ([]int64, bool, error) {
			return nil, false, expectedErr
		},
		func(requestedIDs []int64) ([]int64, bool, error) {
			t.Fatal("expected access point narrowing not to run after network unit error")
			return nil, false, nil
		},
	)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestNarrowTopoEventCondition_PropagatesNetworkAreaError(t *testing.T) {
	t.Helper()

	expectedErr := errors.New("network area permission denied")
	_, err := narrowTopoEventCondition(
		&types.TopoEventCondition{},
		func(requestedIDs []int64) ([]int64, bool, error) {
			return nil, false, expectedErr
		},
		func(requestedIDs []int64) ([]int64, bool, error) {
			t.Fatal("expected network unit narrowing not to run after network area error")
			return nil, false, nil
		},
		func(requestedIDs []int64) ([]int64, bool, error) {
			t.Fatal("expected access point narrowing not to run after network area error")
			return nil, false, nil
		},
	)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}
