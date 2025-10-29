/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package base

import (
	"context"
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// FieldKeyTestDataID defines the key of data id field.
	FieldKeyTestDataID = "data.id"

	// FieldKeyTestDataName defines the key of data name field.
	FieldKeyTestDataName = "data.name"

	// FieldKeyTestDataEnabled defines the key of data enabled field.
	FieldKeyTestDataEnabled = "data.enabled"

	// FieldKeyTestDataValue defines the key of data value field.
	FieldKeyTestDataValue = "data.value"
)

// TestData represents a test data structure for ORM testing
type TestData struct {
	ID      string `json:"id" bson:"id"`
	Name    string `json:"name" bson:"name"`
	Value   int    `json:"value" bson:"value"`
	Enabled bool   `json:"enabled" bson:"enabled"`
}

// UniqueKey returns the unique key for the test data
func (d *TestData) UniqueKey() string {
	return d.ID
}

// UniqueFields returns the unique fields for the test data
func (d *TestData) UniqueFields() []string {
	return []string{FieldKeyTestDataID}
}

// TestDao implements IDao interface for testing
type TestDao struct {
	client     *mongo.Client
	collection *mongo.Collection
	tableName  string
}

// NewTestDao creates a new test DAO
func NewTestDao(client *mongo.Client, databaseName string) *TestDao {
	tableName := "test_orm_" + uuid.NewString()[:8]
	collection := client.Database(databaseName).Collection(tableName)

	return &TestDao{
		client:     client,
		collection: collection,
		tableName:  tableName,
	}
}

// GetClient returns the MongoDB collection
func (d *TestDao) GetClient() *mongo.Collection {
	return d.collection
}

// GetTableName returns the table name
func (d *TestDao) GetTableName() string {
	return d.tableName
}

// GetIndexes returns the indexes for the collection
func (d *TestDao) GetIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{{Key: FieldKeyTestDataName, Value: 1}},
		},
	}
}

// testClient creates a test ORM instance
func testClient(t *testing.T) (*Orm[*TestData, TestData], *TestDao) {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	nCtx := context.Background()
	mongoClient, err := mongo.Connect(
		nCtx,
		&options.ClientOptions{
			Hosts: []string{
				os.Getenv("MONGO_ADDRESS"),
			},
			Auth: &options.Credential{
				Username:      os.Getenv("MONGO_USER"),
				Password:      os.Getenv("MONGO_PASSWORD"),
				AuthSource:    os.Getenv("MONGO_AUTH_SOURCE"),
				AuthMechanism: os.Getenv("MONGO_AUTH_MECHANISM"),
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	databaseName := os.Getenv("MONGO_DATABASE")
	if databaseName == "" {
		databaseName = "test"
	}

	testDao := NewTestDao(mongoClient, databaseName)
	orm := NewOrm[*TestData, TestData](testDao)

	// Clean up test data before running tests
	cleanupTestData(t, testDao)

	// Ensure indexes (after cleanup to avoid conflicts)
	err = orm.EnsureIndexes()
	if err != nil {
		// If it's an index already exists error, we can continue
		if !isIndexAlreadyExistsError(err) {
			t.Fatal(err)
		}
	}

	return orm, testDao
}

// isIndexAlreadyExistsError checks if the error is due to an index already existing
func isIndexAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return errStr == "IndexAlreadyExists" ||
		errStr == "(IndexAlreadyExists) Identical index already exists: id_1" ||
		errStr == "(IndexAlreadyExists) Identical index already exists: data.id_1"
}

// cleanupTestData cleans up all test data
func cleanupTestData(t *testing.T, dao *TestDao) {
	nCtx := contextx.New(context.Background())

	// Drop the entire collection to ensure clean test environment
	err := dao.GetClient().Drop(nCtx)
	if err != nil {
		t.Logf("Warning: failed to drop collection: %v", err)
	}
}

// TestOrm_Create tests the Create method
func TestOrm_Create(t *testing.T) {
	orm, _ := testClient(t)
	nCtx := contextx.New(context.Background())

	tests := []struct {
		name    string
		nCtx    contextx.IContext
		data    *TestData
		wantErr bool
	}{
		{
			name: "create normal data",
			nCtx: nCtx,
			data: &TestData{
				ID:      uuid.NewString(),
				Name:    "test-data-1",
				Value:   100,
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "create data with duplicate ID",
			nCtx: nCtx,
			data: &TestData{
				ID:      "duplicate-id",
				Name:    "test-data-2",
				Value:   200,
				Enabled: false,
			},
			wantErr: false, // Should not error on first create
		},
		{
			name: "nil context",
			nCtx: nil,
			data: &TestData{
				ID:      uuid.NewString(),
				Name:    "test-data-3",
				Value:   300,
				Enabled: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := orm.Create(tt.nCtx, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Test duplicate ID creation
	t.Run("create duplicate ID should fail", func(t *testing.T) {
		duplicateData := &TestData{
			ID:      "duplicate-id",
			Name:    "test-data-duplicate",
			Value:   400,
			Enabled: true,
		}

		err := orm.Create(nCtx, duplicateData)
		if err == nil {
			t.Error("failed to create duplicate ID")
		}
	})
}

// TestOrm_Get tests the Get method
func TestOrm_Get(t *testing.T) {
	orm, _ := testClient(t)
	nCtx := contextx.New(context.Background())

	// Create test data first
	testID := uuid.NewString()
	testData := &TestData{
		ID:      testID,
		Name:    "test-get-data",
		Value:   500,
		Enabled: true,
	}

	err := orm.Create(nCtx, testData)
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	tests := []struct {
		name    string
		nCtx    contextx.IContext
		filter  bson.D
		fields  []string
		wantErr bool
	}{
		{
			name:    "get by ID",
			nCtx:    nCtx,
			filter:  bson.D{{Key: FieldKeyTestDataID, Value: testID}},
			fields:  []string{},
			wantErr: false,
		},
		{
			name:    "get by name",
			nCtx:    nCtx,
			filter:  bson.D{{Key: FieldKeyTestDataName, Value: "test-get-data"}},
			fields:  []string{},
			wantErr: false,
		},
		{
			name:    "get with specific fields",
			nCtx:    nCtx,
			filter:  bson.D{{Key: FieldKeyTestDataID, Value: testID}},
			fields:  []string{FieldKeyTestDataID, FieldKeyTestDataName},
			wantErr: false,
		},
		{
			name:    "get non-existent data",
			nCtx:    nCtx,
			filter:  bson.D{{Key: FieldKeyTestDataID, Value: "non-existent-id"}},
			fields:  []string{},
			wantErr: true,
		},
		{
			name:    "nil context",
			nCtx:    nil,
			filter:  bson.D{{Key: FieldKeyTestDataID, Value: testID}},
			fields:  []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := orm.Get(tt.nCtx, tt.filter, tt.fields...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got == nil {
				t.Error("Get() returned nil data without error")
			}

			if !tt.wantErr && got != nil {
				if got.ID != testID {
					t.Errorf("Get() got ID = %v, want %v", got.ID, testID)
				}
			}
		})
	}
}

// TestOrm_Exist tests the Exist method
func TestOrm_Exist(t *testing.T) {
	orm, _ := testClient(t)
	nCtx := contextx.New(context.Background())

	// Create test data first
	testID := uuid.NewString()
	testData := &TestData{
		ID:      testID,
		Name:    "test-exist-data",
		Value:   600,
		Enabled: true,
	}

	err := orm.Create(nCtx, testData)
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	tests := []struct {
		name    string
		nCtx    contextx.IContext
		filter  bson.D
		want    bool
		wantErr bool
	}{
		{
			name:    "exist by ID",
			nCtx:    nCtx,
			filter:  bson.D{{Key: FieldKeyTestDataID, Value: testID}},
			want:    true,
			wantErr: false,
		},
		{
			name:    "exist by name",
			nCtx:    nCtx,
			filter:  bson.D{{Key: FieldKeyTestDataName, Value: "test-exist-data"}},
			want:    true,
			wantErr: false,
		},
		{
			name:    "non-existent data",
			nCtx:    nCtx,
			filter:  bson.D{{Key: FieldKeyTestDataID, Value: "non-existent-id"}},
			want:    false,
			wantErr: false,
		},
		{
			name:    "nil context",
			nCtx:    nil,
			filter:  bson.D{{Key: FieldKeyTestDataID, Value: testID}},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := orm.Exist(tt.nCtx, tt.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exist() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("Exist() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestOrm_Count tests the Count method
func TestOrm_Count(t *testing.T) {
	orm, _ := testClient(t)
	nCtx := contextx.New(context.Background())

	// Create test data first
	testData1 := &TestData{
		ID:      uuid.NewString(),
		Name:    "test-count-1",
		Value:   700,
		Enabled: true,
	}
	testData2 := &TestData{
		ID:      uuid.NewString(),
		Name:    "test-count-2",
		Value:   800,
		Enabled: false,
	}

	err := orm.Create(nCtx, testData1)
	if err != nil {
		t.Fatalf("Failed to create test data 1: %v", err)
	}
	err = orm.Create(nCtx, testData2)
	if err != nil {
		t.Fatalf("Failed to create test data 2: %v", err)
	}

	tests := []struct {
		name    string
		nCtx    contextx.IContext
		filter  bson.D
		want    int64
		wantErr bool
	}{
		{
			name:    "count all data",
			nCtx:    nCtx,
			filter:  bson.D{},
			want:    2,
			wantErr: false,
		},
		{
			name:    "count by enabled status",
			nCtx:    nCtx,
			filter:  bson.D{{Key: FieldKeyTestDataEnabled, Value: true}},
			want:    1,
			wantErr: false,
		},
		{
			name:    "count by specific value",
			nCtx:    nCtx,
			filter:  bson.D{{Key: FieldKeyTestDataValue, Value: 700}},
			want:    1,
			wantErr: false,
		},
		{
			name:    "count non-existent data",
			nCtx:    nCtx,
			filter:  bson.D{{Key: "value", Value: 999}},
			want:    0,
			wantErr: false,
		},
		{
			name:    "nil context",
			nCtx:    nil,
			filter:  bson.D{},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := orm.Count(tt.nCtx, tt.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("Count() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("Count() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestOrm_List tests the List method
func TestOrm_List(t *testing.T) {
	orm, _ := testClient(t)
	nCtx := contextx.New(context.Background())

	// Create test data first
	testData1 := &TestData{
		ID:      uuid.NewString(),
		Name:    "test-list-1",
		Value:   900,
		Enabled: true,
	}
	testData2 := &TestData{
		ID:      uuid.NewString(),
		Name:    "test-list-2",
		Value:   1000,
		Enabled: false,
	}
	testData3 := &TestData{
		ID:      uuid.NewString(),
		Name:    "test-list-3",
		Value:   1100,
		Enabled: true,
	}

	err := orm.Create(nCtx, testData1)
	if err != nil {
		t.Fatalf("Failed to create test data 1: %v", err)
	}
	err = orm.Create(nCtx, testData2)
	if err != nil {
		t.Fatalf("Failed to create test data 2: %v", err)
	}
	err = orm.Create(nCtx, testData3)
	if err != nil {
		t.Fatalf("Failed to create test data 3: %v", err)
	}

	tests := []struct {
		name      string
		nCtx      contextx.IContext
		filter    bson.D
		findOpt   *options.FindOptions
		wantCount int
		wantErr   bool
	}{
		{
			name:      "list all data",
			nCtx:      nCtx,
			filter:    bson.D{},
			findOpt:   nil,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "list with limit",
			nCtx:      nCtx,
			filter:    bson.D{},
			findOpt:   options.Find().SetLimit(2),
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "list enabled data",
			nCtx:      nCtx,
			filter:    bson.D{{Key: FieldKeyTestDataEnabled, Value: true}},
			findOpt:   nil,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "list by value range",
			nCtx:      nCtx,
			filter:    bson.D{{Key: FieldKeyTestDataValue, Value: bson.D{{Key: "$gte", Value: 1000}}}},
			findOpt:   nil,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "nil context",
			nCtx:      nil,
			filter:    bson.D{},
			findOpt:   nil,
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := orm.List(tt.nCtx, tt.filter, tt.findOpt)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(got) != tt.wantCount {
				t.Errorf("List() got count = %v, want %v", len(got), tt.wantCount)
			}
		})
	}
}

// TestOrm_CreateMany tests the CreateMany method
func TestOrm_CreateMany(t *testing.T) {
	orm, _ := testClient(t)
	nCtx := contextx.New(context.Background())

	tests := []struct {
		name    string
		nCtx    contextx.IContext
		datas   []*TestData
		wantErr bool
	}{
		{
			name: "create multiple data",
			nCtx: nCtx,
			datas: []*TestData{
				{
					ID:      uuid.NewString(),
					Name:    "batch-1",
					Value:   1200,
					Enabled: true,
				},
				{
					ID:      uuid.NewString(),
					Name:    "batch-2",
					Value:   1300,
					Enabled: false,
				},
			},
			wantErr: false,
		},
		{
			name:    "create empty data",
			nCtx:    nCtx,
			datas:   []*TestData{},
			wantErr: false, // Should not error on empty slice
		},
		{
			name: "nil context",
			nCtx: nil,
			datas: []*TestData{
				{
					ID:      uuid.NewString(),
					Name:    "batch-3",
					Value:   1400,
					Enabled: true,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := orm.CreateMany(tt.nCtx, tt.datas)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateMany() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestOrm_DeleteMany tests the DeleteMany method
func TestOrm_DeleteMany(t *testing.T) {
	orm, _ := testClient(t)
	nCtx := contextx.New(context.Background())

	// Create test data first
	testData1 := &TestData{
		ID:      uuid.NewString(),
		Name:    "delete-1",
		Value:   1500,
		Enabled: true,
	}
	testData2 := &TestData{
		ID:      uuid.NewString(),
		Name:    "delete-2",
		Value:   1600,
		Enabled: false,
	}

	err := orm.Create(nCtx, testData1)
	if err != nil {
		t.Fatalf("Failed to create test data 1: %v", err)
	}
	err = orm.Create(nCtx, testData2)
	if err != nil {
		t.Fatalf("Failed to create test data 2: %v", err)
	}

	tests := []struct {
		name    string
		nCtx    contextx.IContext
		filter  bson.D
		wantErr bool
	}{
		{
			name:    "delete by enabled status",
			nCtx:    nCtx,
			filter:  bson.D{{Key: FieldKeyTestDataEnabled, Value: true}},
			wantErr: false,
		},
		{
			name:    "delete non-existent data",
			nCtx:    nCtx,
			filter:  bson.D{{Key: FieldKeyTestDataValue, Value: 9999}},
			wantErr: false, // Should not error when nothing to delete
		},
		{
			name:    "nil context",
			nCtx:    nil,
			filter:  bson.D{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := orm.DeleteMany(tt.nCtx, tt.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteMany() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Verify deletion (soft delete - count should include deleted records)
	t.Run("verify deletion", func(t *testing.T) {
		count, err := orm.Count(nCtx, bson.D{})
		if err != nil {
			t.Fatalf("Failed to count after deletion: %v", err)
		}
		// Since DeleteMany is soft delete, all 2 documents should still be counted
		if count != 2 {
			t.Errorf("Expected 2 documents after soft deletion, got %v", count)
		}
	})
}
