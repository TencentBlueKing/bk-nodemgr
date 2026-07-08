package systeminfo

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	toolsModuleRootRel     = "../../.."
	systeminfoImportPath   = "github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/systeminfo"
	forbiddenPkgImportPath = "github.com/TencentBlueKing/bk-nodemgr/tools/pkg/systeminfo"
)

type callpointSpec struct {
	name                string
	path                string
	stepArg             string
	firstBusinessAnchor string
}

var installerCallpointSpecs = []callpointSpec{
	{
		name:                "node full install",
		path:                "cmd/installer/node/full_install.go",
		stepArg:             "nodeInstaller.StepGeneral",
		firstBusinessAnchor: "filedownloader.NewStep",
	},
	{
		name:                "node full uninstall",
		path:                "cmd/installer/node/full_uninstall.go",
		stepArg:             "nodeInstaller.StepGeneral",
		firstBusinessAnchor: "nodestopper.NewStep",
	},
	{
		name:                "node full reconfig",
		path:                "cmd/installer/node/full_reconfig.go",
		stepArg:             "node.StepGeneral",
		firstBusinessAnchor: "configfetcher.NewStep",
	},
	{
		name:                "plugin full install",
		path:                "cmd/installer/plugin/full_install.go",
		stepArg:             "pluginInstaller.StepGeneral",
		firstBusinessAnchor: "filedownloader.NewStep",
	},
	{
		name:                "plugin full uninstall",
		path:                "cmd/installer/plugin/full_uninstall.go",
		stepArg:             "pluginInstaller.StepGeneral",
		firstBusinessAnchor: "pluginuninstaller.NewStep",
	},
	{
		name:                "plugin full debug",
		path:                "cmd/installer/plugin/full_debug.go",
		stepArg:             "pluginInstaller.StepGeneral",
		firstBusinessAnchor: "pluginrunner.NewStep",
	},
	{
		name:                "pluginv2 full install",
		path:                "cmd/installer/pluginv2/full_install.go",
		stepArg:             "pluginv2Installer.StepGeneral",
		firstBusinessAnchor: "filedownloader.NewStep",
	},
}

func TestInstallerCallpointsLogInitialTargetInfoBeforeBusinessSteps(t *testing.T) {
	for _, spec := range installerCallpointSpecs {
		t.Run(spec.name, func(t *testing.T) {
			file := parseGoFile(t, toolsPath(spec.path))

			assertImportsPath(t, file, systeminfoImportPath)
			assertDoesNotImportPath(t, file, forbiddenPkgImportPath)

			startPositions := callPositions(file, "lHandler.Start")
			if len(startPositions) == 0 {
				t.Fatalf("%s: missing lHandler.Start() call", spec.path)
			}

			systeminfoCalls := callPositions(file, "systeminfo.LogInitialTargetInfo")
			if len(systeminfoCalls) != 1 {
				t.Fatalf("%s: systeminfo.LogInitialTargetInfo call count = %d, want 1", spec.path, len(systeminfoCalls))
			}
			call := systeminfoCalls[0]
			if len(call.args) != 1 {
				t.Fatalf("%s: systeminfo.LogInitialTargetInfo arg count = %d, want 1", spec.path, len(call.args))
			}
			if got := selectorExprString(call.args[0]); got != spec.stepArg {
				t.Fatalf("%s: systeminfo.LogInitialTargetInfo arg = %q, want %q", spec.path, got, spec.stepArg)
			}

			anchorPositions := callPositions(file, spec.firstBusinessAnchor)
			if len(anchorPositions) == 0 {
				t.Fatalf("%s: missing first business anchor %s", spec.path, spec.firstBusinessAnchor)
			}
			startPos := startPositions[0].pos
			callPos := call.pos
			anchorPos := anchorPositions[0].pos
			if !(startPos < callPos && callPos < anchorPos) {
				t.Fatalf(
					"%s: want lHandler.Start() < systeminfo.LogInitialTargetInfo() < %s, got %d < %d < %d",
					spec.path,
					spec.firstBusinessAnchor,
					startPos,
					callPos,
					anchorPos,
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
