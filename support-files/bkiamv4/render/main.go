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

// IAM V4 migration template rendering CLI tool.
//
// Build:
//
//	go build -o iam-render .
//
// Usage:
//
//	./iam-render -t <templates-dir> -v <vars.yaml> -o <output-dir>
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/renderer/gotemplate"
	"gopkg.in/yaml.v3"
)

const (
	outputDirPerm  = 0750
	outputFilePerm = 0644
)

func main() {
	var (
		templateDir string
		varsPath    string
		outputDir   string
	)

	flag.StringVar(&templateDir, "t", "", "Path to the templates directory")
	flag.StringVar(&varsPath, "v", "", "Path to the variables file (.yaml)")
	flag.StringVar(&outputDir, "o", "", "Path to the output directory")
	flag.Parse()

	if templateDir == "" || outputDir == "" {
		fmt.Fprintln(os.Stderr, "Error: both template dir (-t) and output dir (-o) are required")
		flag.Usage()
		os.Exit(1)
	}

	// Load variables (optional)
	var vars map[string]any
	if varsPath != "" {
		var err error
		vars, err = loadVariables(varsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading variables: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Loaded variables from %s\n", varsPath)
	} else {
		vars = make(map[string]any)
		fmt.Fprintln(os.Stderr, "No variables file specified, using defaults")
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, outputDirPerm); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Render all templates
	if err := renderAll(templateDir, outputDir, vars); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func renderAll(templateDir, outputDir string, vars map[string]any) error {
	entries, err := os.ReadDir(templateDir)
	if err != nil {
		return fmt.Errorf("failed to read template directory: %w", err)
	}

	renderer := gotemplate.New()
	count := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".tpl") {
			continue
		}

		// Read template
		templatePath := filepath.Join(templateDir, name)
		templateContent, err := os.ReadFile(templatePath) // nolint:gosec
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", name, err)
		}

		// Render
		result, err := renderer.Render(string(templateContent), vars)
		if err != nil {
			return fmt.Errorf("failed to render %s: %w", name, err)
		}

		// Write output (remove .tpl suffix)
		outputName := strings.TrimSuffix(name, ".tpl")
		outputPath := filepath.Join(outputDir, outputName)
		//nolint:gosec // Operator-selected output contains public model data, not credentials.
		if err := os.WriteFile(outputPath, []byte(result), outputFilePerm); err != nil {
			return fmt.Errorf("failed to write %s: %w", outputName, err)
		}

		fmt.Fprintf(os.Stderr, "Rendered: %s -> %s\n", name, outputName)
		count++
	}

	fmt.Fprintf(os.Stderr, "Total: %d files rendered\n", count)

	return nil
}

func loadVariables(path string) (map[string]any, error) {
	data, err := os.ReadFile(path) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to read variables file: %w", err)
	}

	var vars map[string]any
	if err := yaml.Unmarshal(data, &vars); err != nil {
		return nil, fmt.Errorf("failed to parse variables file: %w", err)
	}

	return vars, nil
}
