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

package systeminfo

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const (
	toolsModuleRootRel     = "../../.."
	systeminfoImportPath   = "github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/systeminfo"
	forbiddenPkgImportPath = "github.com/TencentBlueKing/bk-nodemgr/tools/pkg/systeminfo"
)

type installerCallpoint struct {
	name string
	path string
}

func TestInstallerCallpointsLogInitialTargetInfoBeforeBusinessSteps(t *testing.T) {
	callpoints := installerCallpointFiles(t)
	if len(callpoints) == 0 {
		t.Fatalf("installer callpoint files count = 0, want at least 1")
	}

	for _, callpoint := range callpoints {
		t.Run(callpoint.name, func(t *testing.T) {
			file := parseGoFile(t, callpoint.path)

			assertImportsPath(t, file, systeminfoImportPath)
			assertDoesNotImportPath(t, file, forbiddenPkgImportPath)

			startPositions := callPositions(file, "lHandler.Start")
			if len(startPositions) == 0 {
				t.Fatalf("%s: missing lHandler.Start() call", callpoint.path)
			}

			systeminfoCalls := callPositions(file, "systeminfo.LogInitialTargetInfo")
			if len(systeminfoCalls) != 1 {
				t.Fatalf("%s: systeminfo.LogInitialTargetInfo call count = %d, want 1", callpoint.path, len(systeminfoCalls))
			}
			call := systeminfoCalls[0]
			if len(call.args) != 1 {
				t.Fatalf("%s: systeminfo.LogInitialTargetInfo arg count = %d, want 1", callpoint.path, len(call.args))
			}
			if got := selectorExprString(call.args[0]); !strings.HasSuffix(got, ".StepGeneral") {
				t.Fatalf("%s: systeminfo.LogInitialTargetInfo arg = %q, want *.StepGeneral", callpoint.path, got)
			}

			businessCall, ok := firstBusinessStepCall(file, startPositions[0].pos)
			if !ok {
				t.Fatalf("%s: missing first business NewStep call", callpoint.path)
			}

			startPos := startPositions[0].pos
			callPos := call.pos
			if !(startPos < callPos && callPos < businessCall.pos) {
				t.Fatalf(
					"%s: want lHandler.Start() < systeminfo.LogInitialTargetInfo() < %s, got %d < %d < %d",
					callpoint.path,
					businessCall.selector,
					startPos,
					callPos,
					businessCall.pos,
				)
			}
		})
	}
}

func TestSysteminfoSourcesDoNotImportForbiddenScope(t *testing.T) {
	for _, path := range systeminfoGoFiles(t) {
		t.Run(path, func(t *testing.T) {
			file := parseGoFile(t, path)
			assertDoesNotImportPath(t, file, forbiddenPkgImportPath)
		})
	}
}

func TestTargetInfoFieldsUseOnlyFixedAllowlist(t *testing.T) {
	allowedKeys := map[string]struct{}{
		"goos":            {},
		"goarch":          {},
		"hostname":        {},
		"current_time":    {},
		"timezone":        {},
		"timezone_offset": {},
		"system_version":  {},
		"kernel_version":  {},
	}

	keys := targetInfoFieldKeys(t, parseGoFile(t, "systeminfo.go"))
	if len(keys) != len(allowedKeys) {
		t.Fatalf("targetInfoFields key count = %d, want %d: %v", len(keys), len(allowedKeys), keys)
	}

	seenKeys := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		if _, ok := allowedKeys[key]; !ok {
			t.Fatalf("targetInfoFields contains forbidden key %q", key)
		}
		if _, ok := seenKeys[key]; ok {
			t.Fatalf("targetInfoFields contains duplicate key %q", key)
		}
		seenKeys[key] = struct{}{}
	}

	for key := range allowedKeys {
		if _, ok := seenKeys[key]; !ok {
			t.Fatalf("targetInfoFields missing allowlisted key %q", key)
		}
	}
}

func TestSysteminfoLoggingPathDoesNotUseGoroutines(t *testing.T) {
	for _, path := range systeminfoGoFiles(t) {
		t.Run(path, func(t *testing.T) {
			file := parseGoFile(t, path)
			ast.Inspect(file, func(node ast.Node) bool {
				goStmt, ok := node.(*ast.GoStmt)
				if ok {
					t.Fatalf("%s: unexpected go statement at token position %d", path, goStmt.Go)
				}
				return true
			})
		})
	}
}

type callOccurrence struct {
	pos  token.Pos
	args []ast.Expr
}

func callPositions(file *ast.File, selector string) []callOccurrence {
	positions := make([]callOccurrence, 0)
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || selectorExprString(call.Fun) != selector {
			return true
		}

		positions = append(positions, callOccurrence{pos: call.Pos(), args: call.Args})
		return true
	})

	return positions
}

type businessStepCall struct {
	pos      token.Pos
	selector string
}

func firstBusinessStepCall(file *ast.File, after token.Pos) (businessStepCall, bool) {
	var first businessStepCall
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || call.Pos() <= after {
			return true
		}

		selector := selectorExprString(call.Fun)
		if !isBusinessStepSelector(selector) {
			return true
		}
		if first.pos == token.NoPos || call.Pos() < first.pos {
			first = businessStepCall{pos: call.Pos(), selector: selector}
		}
		return true
	})

	return first, first.pos != token.NoPos
}

func isBusinessStepSelector(selector string) bool {
	return strings.HasSuffix(selector, ".NewStep") && selector != "statusreporter.NewStep"
}

func installerCallpointFiles(t *testing.T) []installerCallpoint {
	t.Helper()

	root := toolsPath("cmd/installer")
	callpoints := make([]installerCallpoint, 0)
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "full_") || !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}

		file := parseGoFile(t, path)
		if len(callPositions(file, "lHandler.Start")) == 0 {
			return nil
		}

		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		callpoints = append(callpoints, installerCallpoint{name: name, path: path})
		return nil
	}); err != nil {
		t.Fatalf("walk installer cmd dir: %v", err)
	}

	sort.Slice(callpoints, func(i, j int) bool {
		return callpoints[i].path < callpoints[j].path
	})
	return callpoints
}

func selectorExprString(expr ast.Expr) string {
	switch typedExpr := expr.(type) {
	case *ast.Ident:
		return typedExpr.Name
	case *ast.SelectorExpr:
		prefix := selectorExprString(typedExpr.X)
		if prefix == "" {
			return typedExpr.Sel.Name
		}
		return prefix + "." + typedExpr.Sel.Name
	default:
		return ""
	}
}

func assertImportsPath(t *testing.T, file *ast.File, importPath string) {
	t.Helper()

	if !importsPath(file, importPath) {
		t.Fatalf("%s: missing import %q", file.Name.Name, importPath)
	}
}

func assertDoesNotImportPath(t *testing.T, file *ast.File, importPath string) {
	t.Helper()

	if importsPath(file, importPath) {
		t.Fatalf("%s: unexpectedly imports %q", file.Name.Name, importPath)
	}
}

func importsPath(file *ast.File, importPath string) bool {
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		if path == importPath {
			return true
		}
	}

	return false
}

func targetInfoFieldKeys(t *testing.T, file *ast.File) []string {
	t.Helper()

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok || len(valueSpec.Names) != 1 || valueSpec.Names[0].Name != "targetInfoFields" {
				continue
			}
			if len(valueSpec.Values) != 1 {
				t.Fatalf("targetInfoFields value count = %d, want 1", len(valueSpec.Values))
			}

			return compositeFieldKeys(t, valueSpec.Values[0])
		}
	}

	t.Fatalf("targetInfoFields declaration not found")
	return nil
}

func compositeFieldKeys(t *testing.T, expr ast.Expr) []string {
	t.Helper()

	literal, ok := expr.(*ast.CompositeLit)
	if !ok {
		t.Fatalf("targetInfoFields value is %T, want *ast.CompositeLit", expr)
	}

	keys := make([]string, 0, len(literal.Elts))
	for _, element := range literal.Elts {
		fieldLiteral, ok := element.(*ast.CompositeLit)
		if !ok {
			t.Fatalf("targetInfoFields element is %T, want *ast.CompositeLit", element)
		}
		keys = append(keys, targetInfoFieldKey(t, fieldLiteral))
	}

	return keys
}

func targetInfoFieldKey(t *testing.T, literal *ast.CompositeLit) string {
	t.Helper()

	for _, element := range literal.Elts {
		keyValue, ok := element.(*ast.KeyValueExpr)
		if !ok || selectorExprString(keyValue.Key) != "key" {
			continue
		}
		basicLiteral, ok := keyValue.Value.(*ast.BasicLit)
		if !ok || basicLiteral.Kind != token.STRING {
			t.Fatalf("targetInfoFields key value is %T, want string literal", keyValue.Value)
		}
		key, err := strconv.Unquote(basicLiteral.Value)
		if err != nil {
			t.Fatalf("unquote targetInfoFields key %s: %v", basicLiteral.Value, err)
		}

		return key
	}

	t.Fatalf("targetInfoFields element missing key field")
	return ""
}

func systeminfoGoFiles(t *testing.T) []string {
	t.Helper()

	paths := make([]string, 0)
	if err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		paths = append(paths, path)
		return nil
	}); err != nil {
		t.Fatalf("walk systeminfo package dir: %v", err)
	}

	return paths
}

func parseGoFile(t *testing.T, path string) *ast.File {
	t.Helper()

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	return file
}

func toolsPath(path string) string {
	return filepath.Join(toolsModuleRootRel, path)
}
