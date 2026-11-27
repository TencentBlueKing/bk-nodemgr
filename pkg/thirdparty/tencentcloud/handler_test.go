/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package tencentcloud

import (
	"context"
	"errors"
	"testing"

	sdkErrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	sdkVpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/cloudsg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

type mockVPCAPI struct {
	ingress        []*sdkVpc.SecurityGroupPolicy
	created        [][]*sdkVpc.SecurityGroupPolicy
	createErr      error
	describeCalled int
}

func (m *mockVPCAPI) DescribeSecurityGroupPolicies(request *sdkVpc.DescribeSecurityGroupPoliciesRequest) (
	*sdkVpc.DescribeSecurityGroupPoliciesResponse, error) {
	if err := request.GetContext().Err(); err != nil {
		return nil, err
	}
	m.describeCalled++

	ingress := make([]*sdkVpc.SecurityGroupPolicy, len(m.ingress))
	copy(ingress, m.ingress)
	response := sdkVpc.NewDescribeSecurityGroupPoliciesResponse()
	response.Response = &sdkVpc.DescribeSecurityGroupPoliciesResponseParams{
		SecurityGroupPolicySet: &sdkVpc.SecurityGroupPolicySet{Ingress: ingress},
	}

	return response, nil
}

func (m *mockVPCAPI) CreateSecurityGroupPolicies(request *sdkVpc.CreateSecurityGroupPoliciesRequest) (
	*sdkVpc.CreateSecurityGroupPoliciesResponse, error) {
	if err := request.GetContext().Err(); err != nil {
		return nil, err
	}
	if m.createErr != nil {
		return nil, m.createErr
	}

	m.created = append(m.created, request.SecurityGroupPolicySet.Ingress)
	m.ingress = append(m.ingress, request.SecurityGroupPolicySet.Ingress...)

	return sdkVpc.NewCreateSecurityGroupPoliciesResponse(), nil
}

func strPtr(s string) *string {
	return &s
}

func managedPolicy(cidr, port string) *sdkVpc.SecurityGroupPolicy {
	return &sdkVpc.SecurityGroupPolicy{
		CidrBlock:         strPtr(cidr),
		Port:              strPtr(port),
		Protocol:          strPtr(ruleProtocolTCP),
		Action:            strPtr(ruleActionAccept),
		PolicyDescription: strPtr(ManagedRuleDescription),
	}
}

func newTestHandler(api vpcAPI) *securityGroupHandler {
	return &securityGroupHandler{clientFactory: func(secretID, secretKey, region string) (vpcAPI, error) {
		return api, nil
	}}
}

func testPool() SecurityGroupPool {
	return SecurityGroupPool{
		Region:           "ap-test",
		SecretID:         "secret-id",
		SecretKey:        "secret-key",
		SecurityGroupIDs: []string{"sg-1"},
	}
}

func TestAuthorizeAddsMissingRulesOnly(t *testing.T) {
	api := &mockVPCAPI{ingress: []*sdkVpc.SecurityGroupPolicy{managedPolicy("1.2.3.4/32", "28668")}}
	handler := newTestHandler(api)
	rules := []cloudsg.CloudSecurityGroupRule{
		{CIDR: "1.2.3.4/32", Port: "28668", Protocol: "TCP"},
		{CIDR: "5.6.7.8/32", Port: "28668", Protocol: "TCP"},
	}

	if err := handler.Authorize(nil, testPool(), rules); err != nil {
		t.Fatalf("authorize failed: %v", err)
	}

	if len(api.created) != 1 || len(api.created[0]) != 1 {
		t.Fatalf("expected exactly one created rule, got %+v", api.created)
	}
	created := api.created[0][0]
	if *created.CidrBlock != "5.6.7.8/32" {
		t.Fatalf("created wrong rule: %+v", created)
	}
	if *created.PolicyDescription != ManagedRuleDescription {
		t.Fatalf("created rule misses managed marker: %+v", created)
	}
}

func TestAuthorizeAppendsRemarkToDescription(t *testing.T) {
	api := &mockVPCAPI{}
	handler := newTestHandler(api)
	rules := []cloudsg.CloudSecurityGroupRule{
		{CIDR: "1.2.3.4/32", Port: "28668", Protocol: "TCP", Remark: "na=0"},
	}

	if err := handler.Authorize(nil, testPool(), rules); err != nil {
		t.Fatalf("authorize failed: %v", err)
	}

	created := api.created[0][0]
	want := ManagedRuleDescription + "|na=0"
	if *created.PolicyDescription != want {
		t.Fatalf("description = %q, want %q", *created.PolicyDescription, want)
	}
}

func TestInspectRecognizesRulesWithRemark(t *testing.T) {
	policy := managedPolicy("1.2.3.4/32", "28668")
	*policy.PolicyDescription = ManagedRuleDescription + ":na=0"

	api := &mockVPCAPI{ingress: []*sdkVpc.SecurityGroupPolicy{policy}}
	handler := newTestHandler(api)

	inspections, err := handler.Inspect(nil, testPool())
	if err != nil {
		t.Fatalf("inspect failed: %v", err)
	}

	if len(inspections[0].Rules) != 1 {
		t.Fatalf("rule with remark should be recognized as managed, got %+v", inspections[0].Rules)
	}
}

func TestAuthorizeSkipsCoveredRules(t *testing.T) {
	broadPolicy := managedPolicy("1.2.3.0/24", "28668")
	allPortPolicy := managedPolicy("5.6.7.8/32", "ALL")
	api := &mockVPCAPI{ingress: []*sdkVpc.SecurityGroupPolicy{broadPolicy, allPortPolicy}}
	handler := newTestHandler(api)

	rules := []cloudsg.CloudSecurityGroupRule{
		{CIDR: "1.2.3.4/32", Port: "28668", Protocol: "TCP"},
		{CIDR: "5.6.7.8/32", Port: "28625,28668", Protocol: "TCP"},
	}

	if err := handler.Authorize(nil, testPool(), rules); err != nil {
		t.Fatalf("authorize failed: %v", err)
	}

	if len(api.created) != 0 {
		t.Fatalf("covered rules should not be created, got %+v", api.created)
	}
}

func TestAuthorizeSkipsCombinedRuleCoveredByScatteredRules(t *testing.T) {
	api := &mockVPCAPI{ingress: []*sdkVpc.SecurityGroupPolicy{
		managedPolicy("1.2.3.4/32", "28625"),
		managedPolicy("1.2.3.4/32", "28668"),
		managedPolicy("1.2.3.4/32", "28925"),
	}}
	handler := newTestHandler(api)

	rules := []cloudsg.CloudSecurityGroupRule{
		{CIDR: "1.2.3.4/32", Port: "28625,28668,28925", Protocol: "TCP"},
	}

	if err := handler.Authorize(nil, testPool(), rules); err != nil {
		t.Fatalf("authorize failed: %v", err)
	}

	if len(api.created) != 0 {
		t.Fatalf("combined rule covered by scattered rules should not be created, got %+v", api.created)
	}
}

func TestAuthorizeIsIdempotent(t *testing.T) {
	api := &mockVPCAPI{ingress: []*sdkVpc.SecurityGroupPolicy{managedPolicy("1.2.3.4/32", "28668")}}
	handler := newTestHandler(api)
	rules := []cloudsg.CloudSecurityGroupRule{{CIDR: "1.2.3.4/32", Port: "28668", Protocol: "TCP"}}

	for round := 0; round < 2; round++ {
		if err := handler.Authorize(nil, testPool(), rules); err != nil {
			t.Fatalf("authorize round %d failed: %v", round, err)
		}
	}

	if len(api.created) != 0 {
		t.Fatalf("idempotent authorize should create nothing, got %+v", api.created)
	}
}

func TestAuthorizeQuotaExceeded(t *testing.T) {
	api := &mockVPCAPI{createErr: sdkErrors.NewTencentCloudSDKError("LimitExceeded", "quota", "req-1")}
	handler := newTestHandler(api)

	err := handler.Authorize(nil, testPool(), []cloudsg.CloudSecurityGroupRule{
		{CIDR: "1.2.3.4/32", Port: "28668", Protocol: "TCP"},
	})

	quotaErr := new(*QuotaExceededError)
	if !errors.As(err, quotaErr) {
		t.Fatalf("expected QuotaExceededError, got %v", err)
	}
}

func TestInspectReturnsAllRules(t *testing.T) {
	userPolicy := managedPolicy("9.9.9.9/32", "22")
	*userPolicy.PolicyDescription = "ops-manual"
	*userPolicy.Action = "DROP"
	api := &mockVPCAPI{ingress: []*sdkVpc.SecurityGroupPolicy{
		managedPolicy("1.2.3.4/32", "28668"),
		userPolicy,
	}}
	handler := newTestHandler(api)

	inspections, err := handler.Inspect(nil, testPool())
	if err != nil {
		t.Fatalf("inspect failed: %v", err)
	}
	if len(inspections) != 1 {
		t.Fatalf("expected one inspection, got %+v", inspections)
	}
	if inspections[0].TotalIngress != 2 {
		t.Fatalf("expected total ingress 2, got %d", inspections[0].TotalIngress)
	}
	// every live rule is reported, user rules are not distinguished from platform rules.
	if len(inspections[0].Rules) != 2 {
		t.Fatalf("expected all rules, got %+v", inspections[0].Rules)
	}
	if inspections[0].Rules[1].Action != "DROP" {
		t.Fatalf("expected the DROP action to be preserved, got %+v", inspections[0].Rules[1])
	}
}

func TestAuthorizeCanceledContextDoesNotCallCloud(t *testing.T) {
	api := &mockVPCAPI{}
	handler := newTestHandler(api)
	nCtx, cancel := contextx.WithCancel(contextx.New(context.Background()))
	cancel()

	err := handler.Authorize(nCtx, testPool(), []cloudsg.CloudSecurityGroupRule{
		{CIDR: "1.2.3.4/32", Port: "28668", Protocol: "TCP"},
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if api.describeCalled != 0 || len(api.created) != 0 {
		t.Fatalf("canceled authorize should not call cloud, describe(%d), create(%d)", api.describeCalled, len(api.created))
	}
}
