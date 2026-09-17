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
Return the proper Tempo image name.
*/}}
{{- define "bk-nodemgr.tempo.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.tempo.image "global" .Values.global) }}
{{- end -}}

{{/*
Return the proper Tempo image registry secret names.
*/}}
{{- define "bk-nodemgr.tempo.imagePullSecrets" -}}
{{ include "common.images.pullSecrets" (dict "images" (list .Values.tempo.image) "global" .Values.global) }}
{{- end -}}

{{/*
Return the embedded Tempo OTLP gRPC endpoint.
*/}}
{{- define "bk-nodemgr.tempo.otlpGrpcEndpoint" -}}
{{ template "bk-nodemgr.fullname" . }}-tempo:{{ .Values.tempo.service.ports.otlpGrpc }}
{{- end -}}

{{/*
Return the proper OpenTelemetry Gateway image name.
*/}}
{{- define "bk-nodemgr.opentelemetryGateway.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.opentelemetryGateway.image "global" .Values.global) }}
{{- end -}}

{{/*
Return the proper OpenTelemetry Gateway image registry secret names.
*/}}
{{- define "bk-nodemgr.opentelemetryGateway.imagePullSecrets" -}}
{{ include "common.images.pullSecrets" (dict "images" (list .Values.opentelemetryGateway.image) "global" .Values.global) }}
{{- end -}}

{{/*
Return the embedded OpenTelemetry Gateway OTLP gRPC endpoint.
*/}}
{{- define "bk-nodemgr.opentelemetryGateway.otlpGrpcEndpoint" -}}
{{ template "bk-nodemgr.fullname" . }}-opentelemetry-gateway:{{ .Values.opentelemetryGateway.service.ports.otlpGrpc }}
{{- end -}}

{{/*
Render the OpenTelemetry Gateway config. User config fully replaces defaults.
*/}}
{{- define "bk-nodemgr.opentelemetryGateway.config" -}}
{{- if .Values.opentelemetryGateway.config -}}
{{- include "common.tplvalues.render" (dict "value" .Values.opentelemetryGateway.config "context" $) -}}
{{- else -}}
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:{{ .Values.opentelemetryGateway.service.ports.otlpGrpc }}
      http:
        endpoint: 0.0.0.0:{{ .Values.opentelemetryGateway.service.ports.otlpHttp }}
processors:
  memory_limiter:
    limit_mib: 512
    spike_limit_mib: 128
    check_interval: 5s
  batch: {}
exporters:
  {{- if .Values.tempo.enabled }}
  otlp/tempo:
    endpoint: {{ include "bk-nodemgr.tempo.otlpGrpcEndpoint" . }}
    tls:
      insecure: true
    sending_queue:
      enabled: true
    retry_on_failure:
      enabled: true
  {{- else }}
  debug:
    verbosity: basic
  {{- end }}
service:
  pipelines:
    traces:
      receivers:
        - otlp
      processors:
        - memory_limiter
        - batch
      exporters:
        {{- if .Values.tempo.enabled }}
        - otlp/tempo
        {{- else }}
        - debug
        {{- end }}
{{- end -}}
{{- end -}}

{{/*
Render service tracing config and default empty OTLP endpoints to the in-chart tracing receiver when enabled.
*/}}
{{- define "bk-nodemgr.tracingConfig" -}}
{{- $config := deepCopy .config -}}
{{- if and .context.Values.opentelemetryGateway.enabled (not (get $config "otlpEndpoint")) -}}
{{- $_ := set $config "exporterType" "otlp" -}}
{{- $_ := set $config "otlpEndpoint" (include "bk-nodemgr.opentelemetryGateway.otlpGrpcEndpoint" .context) -}}
{{- $_ := set $config "otlpProtocol" "grpc" -}}
{{- $_ := set $config "otlpInsecure" true -}}
{{- else if and .context.Values.tempo.enabled (not (get $config "otlpEndpoint")) -}}
{{- $_ := set $config "exporterType" "otlp" -}}
{{- $_ := set $config "otlpEndpoint" (include "bk-nodemgr.tempo.otlpGrpcEndpoint" .context) -}}
{{- $_ := set $config "otlpProtocol" "grpc" -}}
{{- $_ := set $config "otlpInsecure" true -}}
{{- end -}}
{{- omit $config "instanceID" | toYaml -}}
{{- end -}}

{{/*
Return the proper bk-nodemgr replica count
{{ include "bk-nodemgr.replicaCount" ( dict "module" .Values.path.to.module ) }}
*/}}
{{- define "bk-nodemgr.replicaCount" -}}
{{- if gt (float64 .root.replicaCount) 1.0 }}
{{- .root.replicaCount -}}
{{- else -}}
{{- default .root.replicaCount .module.replicaCount -}}
{{- end -}}
{{- end -}}

{{/*
Return the proper etcd dependency full name.
*/}}
{{- define "bk-nodemgr.etcd.fullname" -}}
{{- include "common.names.dependency.fullname" (dict "chartName" "etcd" "chartValues" .Values.etcd "context" $) -}}
{{- end -}}

{{/*
Return the proper etcd client port.
*/}}
{{- define "bk-nodemgr.etcd.clientPort" -}}
{{- coalesce .Values.etcd.service.ports.client .Values.etcd.service.port -}}
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
