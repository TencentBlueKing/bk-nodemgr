/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package credit

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/google/uuid"
)

const hostCreditExpiredInterval = time.Hour * 24

func generateHostCreditExpiredAt() time.Time {
	return time.Now().Add(hostCreditExpiredInterval)
}

func generateHostCreditID() string {
	return fmt.Sprintf("host_%d_%s", time.Now().Unix(), uuid.New().String())
}

// CreateHostCredit create host credit.
func (s *Storage) CreateHostCredit(
	ctx contextx.ITenantContext,
	creditData []byte,
) (string, error) {
	creditID := generateHostCreditID()

	encryptedCreditData, err := s.crypter.Encrypt(creditData)
	if err != nil {
		return "", err
	}

	if err := s.daoCredit.Upsert(ctx, creditID, encryptedCreditData, generateHostCreditExpiredAt()); err != nil {
		return "", err
	}

	return creditID, nil
}

// LoadHostCredit load host credit.
func (s *Storage) LoadHostCredit(
	ctx contextx.ITenantContext,
	creditID string,
) ([]byte, error) {

	encryptedCreditData, err := s.daoCredit.Get(ctx, creditID)
	if err != nil {
		return nil, err
	}

	creditData, err := s.crypter.Decrypt(encryptedCreditData)
	if err != nil {
		return nil, err
	}

	return creditData, nil
}

// CheckHostCreditValid check host credit valid.
func (s *Storage) CheckHostCreditValid(
	ctx contextx.ITenantContext,
	creditIDList ...string,
) (map[string]bool, error) {

	return s.daoCredit.CheckValid(ctx, creditIDList...)
}
