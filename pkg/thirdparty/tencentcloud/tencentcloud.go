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
	"strings"

	sdkCommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	sdkErrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	sdkProfile "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sdkVpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/cloudsg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

// vpcAPI is the subset of the Tencent VPC SDK client used by the handler.
type vpcAPI interface {
	DescribeSecurityGroupPolicies(request *sdkVpc.DescribeSecurityGroupPoliciesRequest) (
		*sdkVpc.DescribeSecurityGroupPoliciesResponse, error)
	CreateSecurityGroupPolicies(request *sdkVpc.CreateSecurityGroupPoliciesRequest) (
		*sdkVpc.CreateSecurityGroupPoliciesResponse, error)
}

type vpcClientFactory func(secretID, secretKey, region string) (vpcAPI, error)

type securityGroupHandler struct {
	clientFactory vpcClientFactory
}

var _ ISecurityGroupHandler = &securityGroupHandler{}

func newVPCClient(secretID, secretKey, region string) (vpcAPI, error) {
	return sdkVpc.NewClientWithSecretId(secretID, secretKey, region)
}

// client builds the vpc api client of one pool, honoring its optional custom endpoint.
func (handler *securityGroupHandler) client(pool SecurityGroupPool) (vpcAPI, error) {
	if pool.Endpoint != "" {
		return newVPCClientWithEndpoint(pool.SecretID, pool.SecretKey, pool.Region, pool.Endpoint)
	}

	return handler.clientFactory(pool.SecretID, pool.SecretKey, pool.Region)
}

// newVPCClientWithEndpoint builds a vpc client calling the given custom api endpoint.
func newVPCClientWithEndpoint(secretID, secretKey, region, endpoint string) (vpcAPI, error) {
	credential := sdkCommon.NewCredential(secretID, secretKey)
	clientProfile := sdkProfile.NewClientProfile()
	clientProfile.HttpProfile.Endpoint = endpoint

	return sdkVpc.NewClient(credential, region, clientProfile)
}

// Authorize adds missing rules to every security group in the pool.
func (handler *securityGroupHandler) Authorize(
	nCtx contextx.IContext, pool SecurityGroupPool, rules []cloudsg.CloudSecurityGroupRule,
) error {

	if len(rules) == 0 {
		return nil
	}

	client, err := handler.client(pool)
	if err != nil {
		return err
	}

	for _, securityGroupID := range pool.SecurityGroupIDs {
		if err := handler.authorizeSecurityGroup(nCtx, client, securityGroupID, rules); err != nil {
			return err
		}
	}

	return nil
}

func (handler *securityGroupHandler) authorizeSecurityGroup(
	nCtx contextx.IContext, client vpcAPI, securityGroupID string, rules []cloudsg.CloudSecurityGroupRule,
) error {

	existing, err := describeRules(nCtx, client, securityGroupID)
	if err != nil {
		return err
	}

	missing := make([]*sdkVpc.SecurityGroupPolicy, 0, len(rules))
	for _, rule := range rules {
		// skip permissions that existing entries already grant, whether by exact match,
		// a broader cidr, or an ALL port/protocol rule.
		if cloudsg.CloudSecurityGroupRuleCoveredBy(existing, rule) {
			continue
		}

		missing = append(missing, buildPolicy(rule))
	}

	if len(missing) == 0 {
		return nil
	}

	request := sdkVpc.NewCreateSecurityGroupPoliciesRequest()
	request.SecurityGroupId = &securityGroupID
	request.SecurityGroupPolicySet = &sdkVpc.SecurityGroupPolicySet{Ingress: missing}
	if nCtx != nil {
		request.SetContext(nCtx)
	}

	if _, err := client.CreateSecurityGroupPolicies(request); err != nil {
		if isQuotaExceeded(err) {
			return &QuotaExceededError{SecurityGroupID: securityGroupID, Err: err}
		}

		return err
	}

	return nil
}

// Inspect queries managed rules and quota usage of every security group in the pool.
func (handler *securityGroupHandler) Inspect(
	nCtx contextx.IContext, pool SecurityGroupPool,
) ([]cloudsg.CloudSecurityGroupInspection, error) {

	client, err := handler.client(pool)
	if err != nil {
		return nil, err
	}

	inspections := make([]cloudsg.CloudSecurityGroupInspection, 0, len(pool.SecurityGroupIDs))
	for _, securityGroupID := range pool.SecurityGroupIDs {
		inspection, err := inspectSecurityGroup(nCtx, client, securityGroupID)
		if err != nil {
			return nil, err
		}

		inspections = append(inspections, inspection)
	}

	return inspections, nil
}

func inspectSecurityGroup(
	nCtx contextx.IContext, client vpcAPI, securityGroupID string,
) (cloudsg.CloudSecurityGroupInspection, error) {

	request := sdkVpc.NewDescribeSecurityGroupPoliciesRequest()
	request.SecurityGroupId = &securityGroupID
	if nCtx != nil {
		request.SetContext(nCtx)
	}

	response, err := client.DescribeSecurityGroupPolicies(request)
	if err != nil {
		return cloudsg.CloudSecurityGroupInspection{}, err
	}

	inspection := cloudsg.CloudSecurityGroupInspection{SecurityGroupID: securityGroupID}
	for _, policy := range response.Response.SecurityGroupPolicySet.Ingress {
		inspection.TotalIngress++

		// the whole pool is managed, every live rule is reported without marker filtering.
		inspection.Rules = append(inspection.Rules, cloudsg.CloudSecurityGroupRule{
			CIDR:     strValue(policy.CidrBlock),
			Port:     strValue(policy.Port),
			Protocol: strValue(policy.Protocol),
			Action:   cloudsg.CloudSecurityGroupRuleAction(strValue(policy.Action)),
		})
	}

	return inspection, nil
}

func describeRules(nCtx contextx.IContext, client vpcAPI, securityGroupID string) ([]cloudsg.CloudSecurityGroupRule, error) {
	request := sdkVpc.NewDescribeSecurityGroupPoliciesRequest()
	request.SecurityGroupId = &securityGroupID
	if nCtx != nil {
		request.SetContext(nCtx)
	}

	response, err := client.DescribeSecurityGroupPolicies(request)
	if err != nil {
		return nil, err
	}

	existing := make([]cloudsg.CloudSecurityGroupRule, 0, len(response.Response.SecurityGroupPolicySet.Ingress))
	for _, policy := range response.Response.SecurityGroupPolicySet.Ingress {
		if strValue(policy.Action) != ruleActionAccept {
			continue
		}

		existing = append(existing, cloudsg.CloudSecurityGroupRule{
			CIDR:     strValue(policy.CidrBlock),
			Port:     strValue(policy.Port),
			Protocol: strValue(policy.Protocol),
			Action:   cloudsg.CloudSecurityGroupRuleAction(strValue(policy.Action)),
		})
	}

	return existing, nil
}

func buildPolicy(rule cloudsg.CloudSecurityGroupRule) *sdkVpc.SecurityGroupPolicy {
	protocol := rule.Protocol
	if protocol == "" {
		protocol = ruleProtocolTCP
	}

	action := ruleActionAccept
	description := ManagedRuleDescription
	if rule.Remark != "" {
		description = description + "|" + rule.Remark
	}

	return &sdkVpc.SecurityGroupPolicy{
		CidrBlock:         &rule.CIDR,
		Port:              &rule.Port,
		Protocol:          &protocol,
		Action:            &action,
		PolicyDescription: &description,
	}
}

func strValue(ptr *string) string {
	if ptr == nil {
		return ""
	}

	return *ptr
}

func isQuotaExceeded(err error) bool {
	sdkErr, ok := err.(*sdkErrors.TencentCloudSDKError)
	if !ok {
		return false
	}

	return strings.Contains(sdkErr.Code, "LimitExceeded") || strings.Contains(sdkErr.Code, "QuotaExceeded")
}
