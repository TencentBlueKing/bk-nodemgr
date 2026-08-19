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

// Package report ...
package logreporter

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func loadTestFile(t *testing.T) {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}
}

// TestReporter_ReportLogs ...
func TestReporter_ReportLogs(t *testing.T) {
	loadTestFile(t)

	file, _ := os.Open(os.Getenv("NODEMGR_LOG_FILE"))

	type fields struct {
		token         string
		reader        io.ReadCloser
		logRptCnt     uint
		bulkSize      int
		reportLogURLs []string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			fields: fields{
				token:         os.Getenv("BK_NODEMGR_TOKEN"),
				reader:        file,
				logRptCnt:     0,
				bulkSize:      110,
				reportLogURLs: []string{os.Getenv("BK_NODEMGR_CALLBACK_ENDPOINT")},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewReporter(ReportLogsArgs{
				Token:         tt.fields.token,
				Reader:        tt.fields.reader,
				LogRptCnt:     tt.fields.logRptCnt,
				BulkSize:      tt.fields.bulkSize,
				ReportLogURLs: tt.fields.reportLogURLs,
			})
			got, err := reporter.ReportLogs(tt.args.ctx)
			if err != nil {
				t.Logf("err:%s", err.Error())
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("ReportLogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.fields.logRptCnt {
				t.Errorf("ReportLogs() got = %v, want %v", got, tt.fields.logRptCnt)
			}
		})
	}
}
