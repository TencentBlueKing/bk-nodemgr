{{/*
Expand the name of the chart.
*/}}
{{- define "mock-server.name" -}}
{{- include "common.names.name" . -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "mock-server.fullname" -}}
{{- include "common.names.fullname" . -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "mock-server.chart" -}}
{{- include "common.names.chart" . -}}
{{- end -}}

{{/*
Return the proper mock-server image name.
*/}}
{{- define "mock-server.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.image "global" .Values.global) }}
{{- end -}}

{{/*
Return the proper Docker Image Registry Secret Names.
*/}}
{{- define "mock-server.imagePullSecrets" -}}
{{ include "common.images.pullSecrets" (dict "images" (list .Values.image) "global" .Values.global) }}
{{- end -}}
