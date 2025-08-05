{{/*
Expand the name of the chart.
*/}}
{{- define "bk-nodemgr.name" -}}
{{- include "common.names.name" . -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "bk-nodemgr.fullname" -}}
{{- include "common.names.fullname" . -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "bk-nodemgr.chart" -}}
{{- include "common.names.chart" . -}}
{{- end -}}

{{/*
Return the proper bk-nodemgr image name
*/}}
{{- define "bk-nodemgr.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.image "global" .Values.global) }}
{{- end -}}

{{/*
Return the proper bk-nodemgr image registry secret names
*/}}
{{- define "bk-nodemgr.imagePullSecrets" -}}
{{ include "common.images.pullSecrets" (dict "images" (list .Values.image) "global" .Values.global) }}
{{- end -}}

{{/*
Return the proper bk-nodemgr replica count
{{ include "bk-nodemgr.replicaCount" ( dict "module" .Values.path.to.module ) }}
*/}}
{{- define "bk-nodemgr.replicaCount" -}}
{{- if gt .root.replicaCount 1.0 }}
{{- .root.replicaCount -}}
{{- else -}}
{{- .module.replicaCount -}}
{{- end -}}
{{- end -}}