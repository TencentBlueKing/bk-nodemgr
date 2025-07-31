/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package messagetracke provides the message storage implementation for relay operations.
package messagetracke

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	filePrefix = "bkmgr_relay_"
	fileSuffix = ".msg"
)

// FileManager manages file storage for messages and implements MessageStore interface.
type FileManager struct {
	storagePath string
	messageSet  map[string]struct{} // mark processed messages.
	ackedSet    map[string]struct{}
	mutex       sync.RWMutex
}

// NewFileManager creates a new FileManager with the specified storage path.
func NewFileManager(storagePath string) MessageTracker {
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		if err := os.MkdirAll(storagePath, 0750); err != nil { //nolint: mnd
			return nil
		}
	}
	fm := &FileManager{
		storagePath: storagePath,
		messageSet:  make(map[string]struct{}),
		ackedSet:    make(map[string]struct{}),
	}
	fm.loadExistingFiles()

	return fm
}

// loadExistingFiles loads existing files into memory set.
func (fm *FileManager) loadExistingFiles() {
	files, _ := filepath.Glob(filepath.Join(fm.storagePath, filePrefix+"*"+fileSuffix))

	for _, file := range files {
		filename := filepath.Base(file)
		idStart := strings.LastIndex(filename, "_") + 1
		messageID := strings.TrimSuffix(filename[idStart:], fileSuffix)

		fm.mutex.Lock()
		fm.messageSet[messageID] = struct{}{}
		fm.mutex.Unlock()
	}
}

// MarkedAcked marks a message ID as acknowledged (in-memory only).
func (fm *FileManager) MarkedAcked(_ context.Context, mid string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	fm.ackedSet[mid] = struct{}{}

	return nil
}

// IsAcked checks if a message ID has been acknowledged.
func (fm *FileManager) IsAcked(_ context.Context, mid string) (bool, error) {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()
	_, exists := fm.ackedSet[mid]

	return exists, nil
}

// MarkProcessed marks a message ID as processed.
func (fm *FileManager) MarkProcessed(_ context.Context, mid string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if _, exists := fm.messageSet[mid]; exists {
		return nil
	}
	fm.messageSet[mid] = struct{}{}

	filename := fm.generateFilename(mid)
	filePath := filepath.Join(fm.storagePath, filename)

	content := []byte(mid + "\n")

	return os.WriteFile(filePath, content, 0600) //nolint: mnd
}

// IsProcessed checks if a message ID has been processed.
func (fm *FileManager) IsProcessed(_ context.Context, mid string) (bool, error) {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()
	_, exists := fm.messageSet[mid]

	return exists, nil
}

// CleanupExpired deletes files older than 24 hours.
func (fm *FileManager) CleanupExpired(_ context.Context) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)
	files, err := filepath.Glob(filepath.Join(fm.storagePath, filePrefix+"*"+fileSuffix))
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := filepath.Base(file)
		parts := strings.Split(filename, "_")
		if len(parts) < 4 { //nolint: mnd
			continue
		}

		timePart := parts[2]
		loc, _ := time.LoadLocation("Local")
		fileTime, err := time.ParseInLocation("20060102-1504", timePart, loc)

		idPart := parts[3]
		messageID := strings.TrimSuffix(idPart, fileSuffix)

		if err == nil && fileTime.Before(cutoff) {
			_ = os.Remove(file)
			delete(fm.messageSet, messageID)
			delete(fm.ackedSet, messageID) // Remove from acked set as well
		}
	}

	return nil
}

// generateFilename generates a filename for the message based on the current timestamp and message ID.
func (fm *FileManager) generateFilename(messageID string) string {
	timestamp := time.Now().Format("20060102-1504")
	return filePrefix + fmt.Sprintf("%s_%s", timestamp, messageID) + fileSuffix
}
