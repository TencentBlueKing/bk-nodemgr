/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package server

import (
	"fmt"
	"io"
	"net/http"

	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
)

// RequestBody rest request body.
type RequestBody interface {
	// Validate validate request body.
	Validate() error

	// AutoConvert auto convert some fields in request body.
	AutoConvert()
}

// Response standard response.
type Response struct {
	Code       resterrf.Code `json:"code"`
	Message    string        `json:"message,omitempty"`
	RequestID  string        `json:"request_id"`
	Data       interface{}   `json:"data,omitempty"`
	Error      *Error        `json:"error,omitempty"`
	Permission *Permission   `json:"permission,omitempty"`
}

// Error defines the error struct.
type Error struct {
	System  string   `json:"system"`
	Message string   `json:"message"`
	Details []Detail `json:"details,omitempty"`
}

// Detail defines the error detail.
type Detail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Permission defines the permission struct.
type Permission struct {
	System     string   `json:"system"`
	SystemName string   `json:"system_name"`
	ApplyURL   string   `json:"apply_url,omitempty"`
	Actions    []Action `json:"actions"`
}

// Action defines the action struct for IAM permission apply response.
type Action struct {
	// ID is the IAM action ID.
	ID string `json:"id"`
	// Name is the display name of the action.
	Name string `json:"name"`
	// RelatedResourceTypes lists the resource types this action operates on.
	RelatedResourceTypes []RelatedResourceType `json:"related_resource_types"`
}

// RelatedResourceType describes a resource type associated with an IAM action.
type RelatedResourceType struct {
	// SystemID is the IAM system that owns this resource type.
	SystemID string `json:"system_id"`
	// Type is the resource type identifier.
	Type string `json:"type"`
	// TypeName is the display name of the resource type.
	TypeName string `json:"type_name"`
}

// PermissionProvider is implemented by errors that carry IAM permission-apply data.
type PermissionProvider interface {
	// PermissionData returns the Permission payload to embed in the response.
	PermissionData() Permission
}

// FileResponse file response.
type FileResponse struct {
	// Data defines file data.
	Data io.ReadCloser

	// Size defines file size.
	Size int64

	// FilePath defines file path.
	FilePath string

	// FileName defines file name.
	FileName string

	// ContentType defines content type.
	ContentType MIMEType

	// Headers defines custom headers.
	Headers map[string]string
}

// StreamResponse stream response.
type StreamResponse struct {
	// Data defines file data.
	Data io.ReadCloser

	// StatusCode defines http status code.
	StatusCode int

	// Headers defines custom headers.
	Headers http.Header
}

// MIMEType defines mime type..
type MIMEType string

// String converts the MIMEType to string.
func (m MIMEType) String() string {
	return string(m)
}

// Validate validates the MIMEType.
func (m MIMEType) Validate() error {
	switch m {
	case MIMETypeBin, MIMETypeText, MIMETypePdf, MIMETypeDoc, MIMETypeXls, MIMETypeZip, MIMETypePng, MIMETypeJpg, MIMETypeJSON:
		return nil
	default:
		return fmt.Errorf("invalid mime type: %s", m)
	}
}

const (
	// MIMETypeJSON defines the mime type of json data.
	MIMETypeJSON MIMEType = "application/json"

	// MIMETypeBin defines the mime type of binary data.
	MIMETypeBin MIMEType = "application/octet-stream"

	// MIMETypeText defines the mime type of text data.
	MIMETypeText MIMEType = "text/plain"

	// MIMETypePdf defines the mime type of pdf data.
	MIMETypePdf MIMEType = "application/pdf"

	// MIMETypeDoc defines the mime type of doc data.
	MIMETypeDoc MIMEType = "application/msword"

	// MIMETypeXls defines the mime type of xls data.
	MIMETypeXls MIMEType = "application/vnd.ms-excel"

	// MIMETypeZip defines the mime type of zip data.
	MIMETypeZip MIMEType = "application/zip"

	// MIMETypePng defines the mime type of png data.
	MIMETypePng MIMEType = "image/png"

	// MIMETypeJpg defines the mime type of jpg data.
	MIMETypeJpg MIMEType = "image/jpeg"
)
