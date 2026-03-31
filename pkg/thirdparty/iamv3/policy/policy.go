// Package policy provides IAM v2 policy expression parsing utilities.
// It transforms iam-go-sdk ExprCell tree into authorized resource ID list.
package policy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/iam-go-sdk/expression"
	"github.com/TencentBlueKing/iam-go-sdk/expression/operator"
)

const (
	// iamFieldSeparator is the separator used by IAM to split resource type and
	// attribute name in policy field expressions (e.g., "biz.id" → type="biz", attr="id").
	iamFieldSeparator = "."

	// iamPathAttribute is the special IAM attribute used for hierarchical resource
	// path conditions. This attribute is not supported by the current parser.
	iamPathAttribute = "_bk_iam_path_"
)

// Parse parses an IAM policy expression and returns the authorized resource scope.
// Returns (true, [], nil) for full access (any), (false, resources, nil) for scoped access,
// or (false, nil, err) for unsupported expressions.
func Parse(expr *expression.ExprCell, systemID, resourceType string) (bool, []types.IAMResource, error) {
	if expr == nil {
		return false, []types.IAMResource{}, nil
	}

	switch expr.OP {
	case operator.Any:
		return true, []types.IAMResource{}, nil
	case operator.Eq:
		return parseEqPolicy(expr, systemID, resourceType)
	case operator.In:
		return parseInPolicy(expr, systemID, resourceType)
	case operator.AND:
		return parseCompositePolicy(expr, systemID, resourceType, true)
	case operator.OR:
		return parseCompositePolicy(expr, systemID, resourceType, false)
	default:
		return false, nil, fmt.Errorf("unsupported policy operation: %s", expr.OP)
	}
}

func parseEqPolicy(expr *expression.ExprCell, systemID, resourceType string) (bool, []types.IAMResource, error) {
	attribute, err := parsePolicyAttribute(expr)
	if err != nil {
		return false, nil, err
	}

	value, ok := expr.Value.(string)
	if !ok {
		return false, nil, fmt.Errorf("policy eq value must be string")
	}

	if attribute != "id" {
		return false, nil, fmt.Errorf("policy contains unsupported attribute condition")
	}

	return false, buildAuthorizedResources(systemID, resourceType, []string{value}), nil
}

func parseInPolicy(expr *expression.ExprCell, systemID, resourceType string) (bool, []types.IAMResource, error) {
	attribute, err := parsePolicyAttribute(expr)
	if err != nil {
		return false, nil, err
	}

	if attribute != "id" {
		return false, nil, fmt.Errorf("policy contains unsupported attribute condition")
	}

	values, ok := expr.Value.([]interface{})
	if !ok {
		return false, nil, fmt.Errorf("policy in value must be array")
	}

	ids := make([]string, 0, len(values))
	for _, value := range values {
		id, ok := value.(string)
		if !ok {
			return false, nil, fmt.Errorf("policy in value must contain strings")
		}
		ids = append(ids, id)
	}

	return false, buildAuthorizedResources(systemID, resourceType, ids), nil
}

func parseCompositePolicy(
	expr *expression.ExprCell,
	systemID, resourceType string,
	isAnd bool,
) (bool, []types.IAMResource, error) {

	if len(expr.Content) == 0 {
		return false, []types.IAMResource{}, nil
	}

	mergedIsAny := false
	mergedResources := []types.IAMResource{}
	for index := range expr.Content {
		childExpr := expr.Content[index]
		childIsAny, childResources, err := Parse(&childExpr, systemID, resourceType)
		if err != nil {
			return false, nil, err
		}

		if index == 0 {
			mergedIsAny = childIsAny
			mergedResources = childResources

			continue
		}

		if isAnd {
			mergedIsAny, mergedResources = intersectAuthorizedInstances(
				mergedIsAny,
				mergedResources,
				childIsAny,
				childResources,
			)
		} else {
			mergedIsAny, mergedResources = unionAuthorizedInstances(
				mergedIsAny,
				mergedResources,
				childIsAny,
				childResources,
			)
		}
	}

	return mergedIsAny, mergedResources, nil
}

func parsePolicyAttribute(expr *expression.ExprCell) (string, error) {
	if expr.Field == "" {
		return "", fmt.Errorf("policy field must be string")
	}

	attribute := getFieldAttribute(expr.Field)
	if attribute == iamPathAttribute {
		return "", fmt.Errorf("policy contains %s, not supported", iamPathAttribute)
	}

	return attribute, nil
}

func unionAuthorizedInstances(
	leftIsAny bool,
	leftResources []types.IAMResource,
	rightIsAny bool,
	rightResources []types.IAMResource,
) (bool, []types.IAMResource) {

	if leftIsAny || rightIsAny {
		return true, []types.IAMResource{}
	}

	resources := append(append([]types.IAMResource{}, leftResources...), rightResources...)

	return false, deduplicateResources(resources)
}

func intersectAuthorizedInstances(
	leftIsAny bool,
	leftResources []types.IAMResource,
	rightIsAny bool,
	rightResources []types.IAMResource,
) (bool, []types.IAMResource) {

	if leftIsAny {
		return rightIsAny, deduplicateResources(rightResources)
	}
	if rightIsAny {
		return leftIsAny, deduplicateResources(leftResources)
	}

	leftSet := make(map[string]types.IAMResource, len(leftResources))
	for _, resource := range leftResources {
		leftSet[resource.ID] = resource
	}

	resources := make([]types.IAMResource, 0)
	for _, resource := range rightResources {
		if matched, ok := leftSet[resource.ID]; ok {
			resources = append(resources, matched)
		}
	}

	return false, deduplicateResources(resources)
}

func deduplicateIDs(ids []string) []string {
	if len(ids) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(ids))
	unique := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}

	sort.Strings(unique)

	return unique
}

func buildAuthorizedResources(systemID, resourceType string, ids []string) []types.IAMResource {
	uniqueIDs := deduplicateIDs(ids)
	resources := make([]types.IAMResource, 0, len(uniqueIDs))
	for _, id := range uniqueIDs {
		resources = append(resources, types.IAMResource{
			SystemID: systemID,
			Type:     resourceType,
			ID:       id,
		})
	}

	return resources
}

func deduplicateResources(resources []types.IAMResource) []types.IAMResource {
	if len(resources) == 0 {
		return []types.IAMResource{}
	}

	seen := make(map[string]struct{}, len(resources))
	unique := make([]types.IAMResource, 0, len(resources))
	for _, resource := range resources {
		key := resource.SystemID + "|" + resource.Type + "|" + resource.ID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, resource)
	}

	sort.Slice(unique, func(left, right int) bool {
		if unique[left].SystemID != unique[right].SystemID {
			return unique[left].SystemID < unique[right].SystemID
		}
		if unique[left].Type != unique[right].Type {
			return unique[left].Type < unique[right].Type
		}

		return unique[left].ID < unique[right].ID
	})

	return unique
}

func getFieldAttribute(field string) string {
	if field == "" {
		return ""
	}

	parts := strings.Split(field, iamFieldSeparator)

	return parts[len(parts)-1]
}
