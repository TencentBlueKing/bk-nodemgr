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

// Package main executes IAM V4 model migrations through API Gateway.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
	"github.com/TencentBlueKing/bk-nodemgr/support-files/bkiamv4/migrate/iamv4"
)

const (
	requestTimeout              = 30 * time.Second
	operationUpsertResourceType = "upsert_resource_type"
	operationUpsertAction       = "upsert_action"
	operationUpsertRole         = "upsert_role"
	fieldName                   = "name"
	fieldID                     = "id"
)

type options struct {
	gatewayURL string
	appCode    string
	appSecret  string
	tenantID   string
	file       string
	directory  string
	dryRun     bool
}

type migration struct {
	SystemID   string      `json:"system_id"`
	Operations []operation `json:"operations"`
	filename   string
}

type operation struct {
	Operation string                     `json:"operation"`
	Data      map[string]json.RawMessage `json:"data"`
	fields    iamv4.SystemFields
	resource  iamv4.ResourceType
	action    iamv4.Action
	role      iamv4.Role
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	var opts options
	flags := flag.NewFlagSet("iam-migrate", flag.ContinueOnError)
	flags.StringVar(&opts.gatewayURL, "gateway-url", "", "Full gateway URL including gateway name and stage")
	flags.StringVar(&opts.appCode, "app-code", "", "Calling BlueKing app code")
	flags.StringVar(&opts.appSecret, "app-secret", os.Getenv("BK_APP_SECRET"), "App secret (defaults to BK_APP_SECRET)")
	flags.StringVar(&opts.tenantID, "tenant-id", "", "Target tenant ID (required)")
	flags.StringVar(&opts.file, "file", "", "Rendered migration JSON file")
	flags.StringVar(&opts.directory, "dir", "", "Directory of numbered migration JSON files")
	flags.BoolVar(&opts.dryRun, "dry-run", false, "Query remote state and print the plan without writing")
	// Do not let flag.PrintDefaults expose the environment-provided secret.
	flags.Lookup("app-secret").DefValue = ""
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if opts.appCode == "" || opts.appSecret == "" || strings.TrimSpace(opts.tenantID) == "" {
		return fmt.Errorf("app-code, app-secret (or BK_APP_SECRET), and tenant-id are required")
	}
	if (opts.file == "") == (opts.directory == "") {
		return fmt.Errorf("exactly one of --file and --dir is required")
	}
	migrations, err := loadMigrations(opts)
	if err != nil {
		return err
	}
	logger.Init(logger.Config{ToStdErr: true})
	defer logger.G.Flush()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	nCtx := contextx.New(ctx, contextx.WithTenantID(opts.tenantID))
	base, err := parseHTTPURL(opts.gatewayURL)
	if err != nil {
		return fmt.Errorf("invalid gateway-url: %w", err)
	}
	endpoints := []string{base.Scheme + "://" + base.Host}
	httpClient, err := restclient.NewHTTPClient(nil)
	if err != nil {
		return fmt.Errorf("create HTTP client: %w", err)
	}
	defer httpClient.CloseIdleConnections()
	// Bound requests and never forward application authentication on redirects.
	httpClient.Timeout = requestTimeout
	httpClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	traceSvc, err := tracing.G().NewService(tracing.ServiceConfig{
		ServiceName: "iam-v4-migrate", ServiceCategory: tracing.ServiceCategoryHTTP,
		SampleRate: 0, // This short-lived tool does not export request traces.
	})
	if err != nil {
		return fmt.Errorf("create client tracing: %w", err)
	}
	defer func() { _ = traceSvc.Shutdown(context.Background()) }()
	handler, err := iamv4.New(&restclient.Capability{
		Name: "iam_v4_migrate", HTTPClient: httpClient,
		Discover: restdiscovery.NewDiscovery("iam-v4-migrate", endpoints), TraceSvc: traceSvc,
	}, &iamv4.Config{
		BaseURL:   strings.TrimRight(base.Path, "/"),
		AppConfig: apigwclient.NewAppConfig(endpoints, opts.appCode, opts.appSecret),
	})
	if err != nil {
		return err
	}

	return executeMigrations(nCtx, handler, migrations, opts.dryRun, os.Stdout)
}

func parseHTTPURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid URL")
	}
	if (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Hostname() == "" ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {

		return nil, fmt.Errorf("expected an HTTP(S) URL without user info, query or fragment")
	}

	return parsed, nil
}

func migrationFiles(opts options) ([]string, error) {
	if opts.file != "" {
		return []string{opts.file}, nil
	}
	entries, err := os.ReadDir(opts.directory)
	if err != nil {
		return nil, fmt.Errorf("read migration directory: %w", err)
	}
	type numberedFile struct {
		path   string
		number uint64
	}
	files := make([]numberedFile, 0)
	pattern := regexp.MustCompile(`^(\d+)_.+\.json$`)
	for _, entry := range entries {
		match := pattern.FindStringSubmatch(entry.Name())
		if entry.IsDir() || match == nil {
			continue
		}
		number, err := strconv.ParseUint(match[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid sequence in %s: %w", entry.Name(), err)
		}
		files = append(files, numberedFile{filepath.Join(opts.directory, entry.Name()), number})
	}
	slices.SortFunc(files, func(left, right numberedFile) int {
		if left.number < right.number {
			return -1
		}
		if left.number > right.number {
			return 1
		}

		return strings.Compare(left.path, right.path)
	})
	if len(files) == 0 {
		return nil, fmt.Errorf("no numbered migration JSON files in %s", opts.directory)
	}
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.path)
	}

	return paths, nil
}

func loadMigrations(opts options) ([]migration, error) {
	files, err := migrationFiles(opts)
	if err != nil {
		return nil, err
	}
	migrations := make([]migration, 0, len(files))
	for _, file := range files {
		item, err := loadMigration(file, opts.appCode)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, item)
	}

	return migrations, nil
}

func loadMigration(file, appCode string) (migration, error) {
	var item migration
	// The local operator selects input via --file or --dir; arbitrary paths are intentional.
	content, err := os.ReadFile(file) //nolint:gosec // G304: trusted CLI input, not a remote-supplied path.
	if err != nil {
		return item, fmt.Errorf("read %s: %w", file, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&item); err != nil {
		return item, fmt.Errorf("parse %s: %w", file, err)
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return item, fmt.Errorf("%s must contain exactly one JSON document", file)
	}
	item.filename = file
	if !regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`).MatchString(item.SystemID) || len(item.Operations) == 0 {
		return item, fmt.Errorf("%s: valid system_id and nonempty operations are required", file)
	}
	for index, op := range item.Operations {
		if err := validateOperation(item.SystemID, op, appCode); err != nil {
			return item, fmt.Errorf("%s operation %d: %w", file, index+1, err)
		}
		data, err := json.Marshal(op.Data)
		if err != nil {
			return item, fmt.Errorf("%s operation %d: encode system fields: %w", file, index+1, err)
		}
		var target any = &item.Operations[index].fields
		if op.Operation == operationUpsertResourceType {
			target = &item.Operations[index].resource
		}
		if op.Operation == operationUpsertAction {
			target = &item.Operations[index].action
		}
		if op.Operation == operationUpsertRole {
			target = &item.Operations[index].role
		}
		if err := json.Unmarshal(data, target); err != nil {
			return item, fmt.Errorf("%s operation %d: decode model fields: %w", file, index+1, err)
		}
	}

	return item, nil
}

func validateOperation(systemID string, op operation, appCode string) error {
	if op.Operation == operationUpsertRole {
		return validateRole(op.Data)
	}
	if op.Operation == operationUpsertAction {
		return validateAction(op.Data)
	}
	if op.Operation == operationUpsertResourceType {
		return validateResourceType(op.Data)
	}
	if op.Operation != "upsert_system" {
		return fmt.Errorf("unsupported operation %q", op.Operation)
	}
	var id string
	if err := json.Unmarshal(op.Data[fieldID], &id); err != nil {
		return fmt.Errorf("data.id is required: %w", err)
	}
	if id != systemID {
		return fmt.Errorf("data.id must equal system_id")
	}
	for field, raw := range op.Data {
		if err := validateSystemField(field, raw, appCode); err != nil {
			return err
		}
	}

	return nil
}

func validateSystemField(field string, raw json.RawMessage, appCode string) error {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fmt.Errorf("data.%s must not be null; omit it to preserve the remote value", field)
	}
	switch field {
	case fieldID, fieldName, "description", "callback_url":
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return fmt.Errorf("data.%s must be a string: %w", field, err)
		}
		if field == fieldName && strings.TrimSpace(value) == "" {
			return fmt.Errorf("data.name must not be empty")
		}
		if field == "callback_url" && value != "" {
			if _, err := parseHTTPURL(value); err != nil {
				return fmt.Errorf("invalid callback_url: %w", err)
			}
		}
	case "clients", "managers":
		return validateSystemMembers(field, raw, appCode)
	default:
		return fmt.Errorf("unknown system field %q", field)
	}

	return nil
}

func validateSystemMembers(field string, raw json.RawMessage, appCode string) error {
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return fmt.Errorf("data.%s must be a string array: %w", field, err)
	}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("data.%s must not contain empty entries", field)
		}
	}
	if field == "clients" && !slices.Contains(values, appCode) {
		return fmt.Errorf("data.clients must include the calling app-code")
	}

	return nil
}

//nolint:gocognit // Keep ordered execution and dry-run state transitions together.
func executeMigrations(ctx contextx.IContext, handler iamv4.IHandler, migrations []migration, dryRun bool, out io.Writer) error {
	// Track virtual creates during dry-run so later operations see the planned system.
	plannedSystems := make(map[string]bool)
	resourceTypes := make(map[string]map[string]iamv4.ResourceType)
	actions := make(map[string]map[string]iamv4.Action)
	roles := make(map[string]map[string]iamv4.Role)
	for _, item := range migrations {
		for index, op := range item.Operations {
			if op.Operation == operationUpsertRole {
				if err := item.executeRole(ctx, handler, op, roles, actions, resourceTypes, dryRun, out); err != nil {
					return fmt.Errorf("%s operation %d: %w", item.filename, index+1, err)
				}

				continue
			}
			if op.Operation == operationUpsertAction {
				if err := item.executeAction(ctx, handler, op, actions, resourceTypes, dryRun, out); err != nil {
					return fmt.Errorf("%s operation %d: %w", item.filename, index+1, err)
				}

				continue
			}
			if op.Operation == operationUpsertResourceType {
				if err := item.executeResourceType(ctx, handler, op, resourceTypes, dryRun, out); err != nil {
					return fmt.Errorf("%s operation %d: %w", item.filename, index+1, err)
				}

				continue
			}
			exists := plannedSystems[item.SystemID]
			if !dryRun || !exists {
				var err error
				exists, err = handler.SystemExists(ctx, item.SystemID)
				if err != nil {
					return fmt.Errorf("%s operation %d: %w", item.filename, index+1, err)
				}
			}
			if err := item.executeOperation(ctx, handler, index, op, exists, dryRun, out); err != nil {
				return err
			}
			if dryRun {
				plannedSystems[item.SystemID] = true
				if !exists {
					// The planned system cannot be queried until it is actually created.
					resourceTypes[item.SystemID] = make(map[string]iamv4.ResourceType)
					actions[item.SystemID] = make(map[string]iamv4.Action)
					roles[item.SystemID] = make(map[string]iamv4.Role)
				}
			}
		}
	}

	return nil
}

func (item migration) executeOperation(
	ctx contextx.IContext, handler iamv4.IHandlerSystem, index int, op operation, exists, dryRun bool, out io.Writer,
) error {

	action := "update_system"
	if !exists {
		action = "create_system"
		if _, ok := op.Data[fieldName]; !ok {
			return fmt.Errorf("%s: data.name is required to create system %s", item.filename, item.SystemID)
		}
		if _, ok := op.Data["clients"]; !ok {
			return fmt.Errorf("%s: data.clients is required to create system %s", item.filename, item.SystemID)
		}
	}
	if _, err := fmt.Fprintf(out, "%s operation %d: %s %s (dry-run=%t)\n", item.filename, index+1, action, item.SystemID, dryRun); err != nil {
		return fmt.Errorf("write migration plan: %w", err)
	}
	if dryRun {
		return nil
	}
	if !exists {
		if err := handler.CreateSystem(ctx, item.SystemID, op.fields); err != nil {
			return fmt.Errorf("%s operation %d: %w", item.filename, index+1, err)
		}

		return nil
	}

	if err := handler.UpdateSystem(ctx, item.SystemID, op.fields); err != nil {
		return fmt.Errorf("%s operation %d: %w", item.filename, index+1, err)
	}

	return nil
}
