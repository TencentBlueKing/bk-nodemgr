/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package manager

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	tgzModeDir  = 0755
	tgzModeFile = 0644
	tgzModeExe  = 0755

	tgzPathNameAny1 = "*~1"
	tgzPathNameAny2 = "*~2"
	tgzPathNameAny3 = "*~3"
)

func isTgzPathNameAny(pathName string) bool {
	return pathName == tgzPathNameAny1 ||
		pathName == tgzPathNameAny2 ||
		pathName == tgzPathNameAny3
}

type tgzWriteRuleDir struct {
	targetFilePath []string
	targetFileMode int64
}

type tgzWriteRuleFile struct {
	sourceFilePath []string
	targetFilePath []string
	targetFileMode int64
}

type tgzWriteRuleStream struct {
	sourceFile io.ReadCloser
	fileRules  []tgzWriteRuleFile
}

type tgzReadRule struct {
	filePath []string
	callback func(path []string, r io.Reader) error
}

// generateTgz takes responsibility for all source and target file to close.
// nolint: funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func generateTgz(
	targetFile io.WriteCloser,
	dirRules []tgzWriteRuleDir,
	streamRules []*tgzWriteRuleStream) (err error) {

	// close source and target file.
	defer func() {
		for _, stream := range streamRules {
			if errClose := stream.sourceFile.Close(); errClose != nil {
				err = errors.Join(err, errClose)
			}
		}
		if errClose := targetFile.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	// target gzip writer.
	gzipWriter := gzip.NewWriter(targetFile)
	defer func() {
		if errClose := gzipWriter.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	// target tar writer.
	tarWriter := tar.NewWriter(gzipWriter)
	defer func() {
		if errClose := tarWriter.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	// write dirs.
	for _, rule := range dirRules {
		if err = tarWriter.WriteHeader(&tar.Header{
			Name:     strings.Join(rule.targetFilePath, "/"),
			Mode:     rule.targetFileMode,
			ModTime:  time.Now(),
			Typeflag: tar.TypeDir,
		}); err != nil {
			return fmt.Errorf("failed to write tar header for directory(%v): %w", rule.targetFilePath, err)
		}
	}

	// source gzip reader.
	for _, stream := range streamRules {
		sourceFile := stream.sourceFile
		fileRules := stream.fileRules

		err = copyFileToTgz(sourceFile, fileRules, tarWriter)
		if err != nil {
			return fmt.Errorf("failed to generate tgz: %w", err)
		}
	}

	return nil
}

func copyFileToTgz(sourceFile io.ReadCloser, fileRules []tgzWriteRuleFile, tarWriter *tar.Writer) error {
	gzipReader, err := gzip.NewReader(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file to tgz: %w", err)
	}

	defer func() {
		if errClose := gzipReader.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	// source tar reader.
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		// trim the heading '.' and '/'
		paths := strings.Split(strings.TrimLeft(header.Name, "./"), "/")

		// validates if the paths match the rules.
		for _, rule := range fileRules {
			if len(paths) != len(rule.sourceFilePath) {
				continue
			}

			matched := true
			mapping := make(map[string]string)
			for idx := range paths {
				if isTgzPathNameAny(rule.sourceFilePath[idx]) {
					mapping[rule.sourceFilePath[idx]] = paths[idx]
					continue
				}

				if rule.sourceFilePath[idx] != paths[idx] {
					matched = false
					break
				}
			}

			if !matched {
				continue
			}

			target := make([]string, len(rule.targetFilePath))
			for idx := range rule.targetFilePath {
				if isTgzPathNameAny(rule.targetFilePath[idx]) {
					var ok bool
					target[idx], ok = mapping[rule.targetFilePath[idx]]

					if !ok {
						return fmt.Errorf("failed to map target path. rule(%v), header(%+v)", rule, header)
					}

					continue
				}

				target[idx] = rule.targetFilePath[idx]
			}

			if err = tarWriter.WriteHeader(&tar.Header{
				Name:     strings.Join(target, "/"),
				Mode:     rule.targetFileMode,
				ModTime:  time.Now(),
				Typeflag: tar.TypeReg,
				Size:     header.Size,
			}); err != nil {
				return fmt.Errorf("failed to write tar header for file(%v): %w", target, err)
			}

			// this copy is only for admin usage, so it's ok to ignore the security check.
			// nolint: gosec
			if _, err = io.Copy(tarWriter, tarReader); err != nil {
				return fmt.Errorf("failed to copy file. origin(%v), target(%v): %w", paths, target, err)
			}
		}
	}

	return nil
}

// checkTgz takes responsibility for source file to close.
// nolint: funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func checkTgz(sourceFile io.ReadCloser, rules []tgzReadRule) (err error) {
	// close source file.
	defer func() {
		if errClose := sourceFile.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	// source gzip reader.
	gzipReader, err := gzip.NewReader(sourceFile)
	if err != nil {
		return err
	}
	defer func() {
		if errClose := gzipReader.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	// source tar reader.
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// trim the heading '.' and '/'
		paths := strings.Split(strings.TrimLeft(header.Name, "./"), "/")

		// validates if the paths match the rules.
		for _, rule := range rules {
			if len(paths) < len(rule.filePath) {
				continue
			}

			matched := true
			for idx := range rule.filePath {
				if isTgzPathNameAny(rule.filePath[idx]) {
					continue
				}

				// support prefix or suffix match
				if rule.filePath[idx] != paths[idx] &&
					!strings.HasPrefix(paths[idx], rule.filePath[idx]) &&
					!strings.HasSuffix(paths[idx], rule.filePath[idx]) {

					matched = false
					break
				}
			}

			if !matched {
				continue
			}

			if err = rule.callback(paths, tarReader); err != nil {
				return err
			}
		}
	}

	return nil
}
