/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package messagetracker provides the message storage implementation for relay operations.
package messagetracker

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
	filePrefix           = "bkmgr_relay_"
	fileSuffix           = ".msg"
	fileRecoveryinterval = 4 * time.Hour
	storageDirMode       = 0750
	processedMarkerMode  = 0600
)

// FileTracker manages file storage for messages and implements MessageStore interface.
type FileTracker struct {
	storagePath string
	messageSet  map[string]struct{} // mark processed messages.
	ackedSet    map[string]struct{}
	mutex       sync.RWMutex
}

// NewFileTracker creates a new FileTracker with the specified storage path.
// nolint: mnd
func NewFileTracker(ctx context.Context, storagePath string) (IMessageTracker, error) {
	_, err := os.Stat(storagePath)
	if err == nil {
		return initializeFileTracker(ctx, storagePath), nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	if mkdirErr := os.MkdirAll(storagePath, storageDirMode); mkdirErr != nil {
		return nil, mkdirErr
	}

	return initializeFileTracker(ctx, storagePath), nil
}

func initializeFileTracker(ctx context.Context, storagePath string) *FileTracker {
	fm := &FileTracker{
		storagePath: storagePath,
		messageSet:  make(map[string]struct{}),
		ackedSet:    make(map[string]struct{}),
	}
	fm.loadExistingFiles()

	go func() {
		ticker := time.NewTicker(fileRecoveryinterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := fm.cleanupExpired(ctx); err != nil {
					return
				}
			}
		}
	}()

	return fm
}

// loadExistingFiles loads existing files into memory set.
func (fm *FileTracker) loadExistingFiles() {
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

// IsAcked checks if a message ID has been acknowledged.
func (fm *FileTracker) IsAcked(_ context.Context, mid string) (bool, error) {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()
	_, exists := fm.ackedSet[mid]

	return exists, nil
}

// MarkAcked marks a message ID as acked.
func (fm *FileTracker) MarkAcked(_ context.Context, mid string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	fm.ackedSet[mid] = struct{}{}

	return nil
}

// TryMarkProcessed tries to mark a message ID as processed.
func (fm *FileTracker) TryMarkProcessed(_ context.Context, mid string) (bool, error) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if _, exists := fm.messageSet[mid]; exists {
		return false, nil
	}

	filename := fm.generateFilename(mid)
	filePath := filepath.Join(fm.storagePath, filename)
	content := []byte(mid + "\n")

	if err := os.MkdirAll(fm.storagePath, storageDirMode); err != nil {
		return false, fmt.Errorf("failed to ensure message tracker storage dir for try mark processed, path(%s): %w", fm.storagePath, err)
	}

	if err := os.WriteFile(filePath, content, processedMarkerMode); err != nil {
		return false, fmt.Errorf("failed to write message tracker marker, path(%s): %w", filePath, err)
	}

	fm.messageSet[mid] = struct{}{}

	return true, nil
}

// cleanupExpired deletes files older than 24 hours.
func (fm *FileTracker) cleanupExpired(_ context.Context) error {
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
		if err != nil {
			continue
		}

		if !fileTime.Before(cutoff) {
			continue
		}

		idPart := parts[3]
		messageID := strings.TrimSuffix(idPart, fileSuffix)

		_ = os.Remove(file)
		delete(fm.messageSet, messageID)
		delete(fm.ackedSet, messageID) // Remove from acked set as well
	}

	return nil
}

// generateFilename generates a filename for the message based on the current timestamp and message ID.
func (fm *FileTracker) generateFilename(messageID string) string {
	timestamp := time.Now().Format("20060102-1504")
	return filePrefix + fmt.Sprintf("%s_%s", timestamp, messageID) + fileSuffix
}
