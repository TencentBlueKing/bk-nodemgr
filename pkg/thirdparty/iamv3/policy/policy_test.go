package policy

import (
	"strings"
	"testing"

	"github.com/TencentBlueKing/iam-go-sdk/expression"
	"github.com/TencentBlueKing/iam-go-sdk/expression/operator"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name         string
		expr         *expression.ExprCell
		systemID     string
		resourceType string
		wantIsAny    bool
		wantIDs      []string
		wantErrPart  string
	}{
		{
			name:         "nil policy returns empty ids",
			expr:         nil,
			systemID:     "bk_nodemgr",
			resourceType: "biz",
			wantIDs:      []string{},
		},
		{
			name:         "any policy",
			expr:         &expression.ExprCell{OP: operator.Any},
			systemID:     "bk_nodemgr",
			resourceType: "biz",
			wantIsAny:    true,
		},
		{
			name:         "eq id policy",
			expr:         &expression.ExprCell{OP: operator.Eq, Field: "biz.id", Value: "1"},
			systemID:     "bk_nodemgr",
			resourceType: "biz",
			wantIDs:      []string{"1"},
		},
		{
			name:         "in id policy",
			expr:         &expression.ExprCell{OP: operator.In, Field: "biz.id", Value: []interface{}{"1", "2", "3"}},
			systemID:     "bk_nodemgr",
			resourceType: "biz",
			wantIDs:      []string{"1", "2", "3"},
		},
		{
			name: "or union policy",
			expr: &expression.ExprCell{OP: operator.OR, Content: []expression.ExprCell{
				{OP: operator.In, Field: "biz.id", Value: []interface{}{"1", "2"}},
				{OP: operator.In, Field: "biz.id", Value: []interface{}{"2", "3"}},
			}},
			systemID:     "bk_nodemgr",
			resourceType: "biz",
			wantIDs:      []string{"1", "2", "3"},
		},
		{
			name: "and intersection with any",
			expr: &expression.ExprCell{OP: operator.AND, Content: []expression.ExprCell{
				{OP: operator.Any},
				{OP: operator.In, Field: "biz.id", Value: []interface{}{"2", "3"}},
			}},
			systemID:     "bk_nodemgr",
			resourceType: "biz",
			wantIDs:      []string{"2", "3"},
		},
		{
			name:         "path attribute not supported",
			expr:         &expression.ExprCell{OP: operator.Eq, Field: "biz._bk_iam_path_", Value: "/foo/"},
			systemID:     "bk_nodemgr",
			resourceType: "biz",
			wantErrPart:  "_bk_iam_path_",
		},
		{
			name:         "non id attribute not supported",
			expr:         &expression.ExprCell{OP: operator.Eq, Field: "biz.name", Value: "foo"},
			systemID:     "bk_nodemgr",
			resourceType: "biz",
			wantErrPart:  "unsupported attribute condition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsAny, gotResources, err := Parse(tt.expr, tt.systemID, tt.resourceType)
			if tt.wantErrPart != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrPart) {
					t.Fatalf("Parse() error = %v, want substring %q", err, tt.wantErrPart)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse() unexpected error: %v", err)
			}
			if gotIsAny != tt.wantIsAny {
				t.Fatalf("Parse().isAny = %v, want %v", gotIsAny, tt.wantIsAny)
			}
			gotIDs := make([]string, 0, len(gotResources))
			for _, resource := range gotResources {
				gotIDs = append(gotIDs, resource.ID)
				if resource.SystemID != tt.systemID {
					t.Fatalf("Parse().resource.SystemID = %q, want %q", resource.SystemID, tt.systemID)
				}
				if resource.Type != tt.resourceType {
					t.Fatalf("Parse().resource.Type = %q, want %q", resource.Type, tt.resourceType)
				}
			}
			if strings.Join(gotIDs, ",") != strings.Join(tt.wantIDs, ",") {
				t.Fatalf("Parse().resourceIDs = %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}
}
