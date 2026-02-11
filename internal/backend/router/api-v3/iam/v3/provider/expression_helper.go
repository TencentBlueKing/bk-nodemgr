/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package provider

import (
	"encoding/json"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/iam-go-sdk/expression"
)

// InstanceForEval wraps a resource instance with its attributes for expression evaluation.
type InstanceForEval struct {
	Instance   ResourceInstance       // The resource instance
	Attributes map[string]interface{} // Attributes for evaluation (includes "id", "display_name", "_bk_iam_path_", etc.)
}

// evalExpressionFilter evaluates IAM policy expressions against resource instances.
// It filters instances based on the expression and applies pagination to the results.
//
// Parameters:
//   - expressionMap: The IAM policy expression as a map (from IAM callback request)
//   - resourceType: The resource type identifier (e.g., "network_area", "package")
//   - instances: List of instances with their attributes for evaluation
//   - page: Pagination parameters
//
// Returns:
//   - *ListInstanceData: Filtered and paginated results with total count
//   - error: Any error during expression evaluation
//
// Behavior:
//   - Empty/nil expression: Returns all instances (no filtering)
//   - Expression with op="any": Returns all instances (SDK handles this)
//   - Valid expression: Filters instances using ExprCell.Eval()
//   - Pagination: Applied AFTER filtering, Count = total matches (pre-pagination)
func evalExpressionFilter(
	expressionMap map[string]interface{},
	resourceType string,
	instances []InstanceForEval,
	page types.Page,
) (*ListInstanceData, error) {

	// Handle empty or nil expression - return all instances
	if len(expressionMap) == 0 {
		return paginateInstances(instances, page), nil
	}

	// Deserialize expression map to ExprCell using JSON round-trip
	// This is consistent with the project's conv.MapToStruct pattern
	exprBytes, err := json.Marshal(expressionMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal expression map: %w", err)
	}

	var exprCell expression.ExprCell
	if err := json.Unmarshal(exprBytes, &exprCell); err != nil {
		return nil, fmt.Errorf("failed to unmarshal expression: %w", err)
	}

	// Filter instances by evaluating expression
	var filtered []InstanceForEval
	for _, inst := range instances {
		// Create ObjectSet for this instance
		objSet := expression.NewObjectSet()
		objSet.Set(resourceType, inst.Attributes)

		// Evaluate expression against this instance
		if exprCell.Eval(objSet) {
			filtered = append(filtered, inst)
		}
	}

	// Apply pagination to filtered results
	return paginateInstances(filtered, page), nil
}

// paginateInstances applies pagination to a list of instances.
// Returns ListInstanceData with Count = total matches and Results = paginated subset.
func paginateInstances(instances []InstanceForEval, page types.Page) *ListInstanceData {
	total := int64(len(instances))

	// Calculate pagination boundaries
	offset := int64(page.Offset)
	limit := int64(page.Limit)

	// Handle offset beyond total
	if offset >= total {
		return &ListInstanceData{
			Count:   total,
			Results: []ResourceInstance{},
		}
	}

	// Calculate end index
	end := offset + limit
	if end > total {
		end = total
	}

	// Extract paginated instances
	results := make([]ResourceInstance, 0, end-offset)
	for i := offset; i < end; i++ {
		results = append(results, instances[i].Instance)
	}

	return &ListInstanceData{
		Count:   total,
		Results: results,
	}
}
