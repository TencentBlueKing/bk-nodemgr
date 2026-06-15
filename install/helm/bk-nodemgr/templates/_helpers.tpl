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
Return the proper bk-nodemgr apigw-sync image registry secret names
*/}}
{{- define "bk-nodemgr.apiManagerImagePullSecrets" -}}
{{ include "common.images.pullSecrets" (dict "images" (list .Values.image .Values.apiManagerImage) "global" .Values.global) }}
{{- end -}}

{{/*
Return the proper bk-nodemgr replica count
{{ include "bk-nodemgr.replicaCount" ( dict "module" .Values.path.to.module ) }}
*/}}
{{- define "bk-nodemgr.replicaCount" -}}
{{- if gt (float64 .root.replicaCount) 1.0 }}
{{- .root.replicaCount -}}
{{- else -}}
{{- .module.replicaCount -}}
{{- end -}}
{{- end -}}

{{/*
Render container lifecycle hooks with a default preStop unless preStop is explicitly configured.
*/}}
{{- define "bk-nodemgr.lifecycleHooks" -}}
{{- $context := .context -}}
{{- $hooks := .hooks | default dict -}}
{{- $defaultHooks := dict "preStop" (dict "exec" (dict "command" (list "/bin/sh" "-c" "sleep 5"))) -}}
{{- if hasKey $hooks "preStop" -}}
{{- include "common.tplvalues.render" (dict "value" $hooks "context" $context) -}}
{{- else -}}
{{- $mergedHooks := mergeOverwrite (deepCopy $defaultHooks) $hooks -}}
{{- include "common.tplvalues.render" (dict "value" $mergedHooks "context" $context) -}}
{{- end -}}
{{- end -}}

{{/*
Render workflow config and inject gracefulShutdownTimeoutSeconds from terminationGracePeriodSeconds / 2 when absent.
*/}}
{{- define "bk-nodemgr.workflowConfig" -}}
{{- $config := deepCopy .config -}}
{{- if not (hasKey $config "gracefulShutdownTimeoutSeconds") -}}
{{- $_ := set $config "gracefulShutdownTimeoutSeconds" (div (int .terminationGracePeriodSeconds) 2) -}}
{{- end -}}
{{- $config | toYaml -}}
{{- end -}}

{{/*
Render an HTTP server config and inject gracefulShutdownTimeoutSec from terminationGracePeriodSeconds / 2 when absent.
*/}}
{{- define "bk-nodemgr.httpServerConfig" -}}
{{- $config := deepCopy .config -}}
{{- if not (hasKey $config "gracefulShutdownTimeoutSec") -}}
{{- $_ := set $config "gracefulShutdownTimeoutSec" (div (int .terminationGracePeriodSeconds) 2) -}}
{{- end -}}
{{- $config | toYaml -}}
{{- end -}}
