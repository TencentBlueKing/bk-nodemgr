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

// evalExpressionFilter evaluates IAM policy expressions (map format) against resource instances.
//
// This function wraps evalExprCellFilter and provides backward compatibility for IAM callback handlers
// that receive policy expressions as map[string]interface{}.
//
// Integration with iam-go-sdk:
//  1. Deserializes IAM policy expression map to expression.ExprCell
//  2. Delegates to evalExprCellFilter for core evaluation logic
//
// Parameters:
//   - expressionMap: IAM policy expression (from callback request filter)
//   - resourceType: Resource type for ObjectSet.Set()
//   - instances: Instances with attributes for evaluation
//   - page: Pagination applied AFTER filtering
//
// Returns:
//   - *ListInstanceData: Filtered results with total count
//   - error: Expression parsing or evaluation error
//
// Behavior:
//   - Empty/nil expression: Returns all instances (no filtering)
//   - Valid expression: Deserializes to ExprCell and delegates to evalExprCellFilter
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

	// Delegate to core evaluation logic
	return evalExprCellFilter(&exprCell, resourceType, instances, page)
}

// evalExprCellFilter evaluates IAM policy expression (ExprCell format) against resource instances.
//
// This is the core evaluation logic that directly accepts expression.ExprCell.
// It integrates with iam-go-sdk expression package to perform policy-based filtering.
//
// IAM Expression Evaluation Flow:
//
//	Policy Expression (ExprCell)
//	  ↓ Eval(ObjectSet)
//	expression.ObjectSet (resource type + attributes)
//	  ↓ GetAttribute("type.field")
//	Attribute Value (e.g., _bk_iam_path_: ["/networkarea,123/"])
//	  ↓ Compare with policy value using operator
//	Boolean Result (true = authorized, false = denied)
//
// Special handling by SDK:
//   - _bk_iam_path_ with starts_with: Strips ",*/" suffix for wildcard matching
//   - Array values: Evaluates each element (any match = true for positive ops)
//   - Logical operators: AND/OR recursively evaluate sub-expressions
//
// Parameters:
//   - exprCell: IAM policy expression (from IAM or auth layer)
//   - resourceType: Resource type for ObjectSet.Set()
//   - instances: Instances with attributes for evaluation
//   - page: Pagination applied AFTER filtering
//
// Returns:
//   - *ListInstanceData: Filtered results with total count
//   - error: Expression evaluation error
//
// Behavior:
//   - nil expression: Returns all instances (no filtering)
//   - Expression with op="any": Returns all instances (SDK handles this)
//   - Valid expression: Filters instances using ExprCell.Eval()
//   - Pagination: Applied AFTER filtering, Count = total matches (pre-pagination)
func evalExprCellFilter(
	exprCell *expression.ExprCell,
	resourceType string,
	instances []InstanceForEval,
	page types.Page,
) (*ListInstanceData, error) {

	// Handle nil expression - return all instances
	if exprCell == nil {
		return paginateInstances(instances, page), nil
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
