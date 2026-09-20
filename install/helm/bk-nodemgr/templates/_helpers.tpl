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
Return the embedded Tempo dependency full name.
*/}}
{{- define "bk-nodemgr.tempo.fullname" -}}
{{- if .Values.tempo.fullnameOverride -}}
{{- .Values.tempo.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default "tempo" .Values.tempo.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Return the embedded Tempo OTLP gRPC endpoint.
*/}}
{{- define "bk-nodemgr.tempo.otlpGrpcEndpoint" -}}
{{- $endpoint := dig "tempo" "receivers" "otlp" "protocols" "grpc" "endpoint" "0.0.0.0:4317" .Values.tempo -}}
{{- $port := regexSplit ":" $endpoint -1 | last -}}
{{ include "bk-nodemgr.tempo.fullname" . }}:{{ $port }}
{{- end -}}

{{/*
Return the OpenTelemetry Gateway full name used by the official collector subchart.
*/}}
{{- define "bk-nodemgr.opentelemetryGateway.fullname" -}}
{{- if .Values.opentelemetryGateway.fullnameOverride -}}
{{- .Values.opentelemetryGateway.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default "opentelemetry-gateway" .Values.opentelemetryGateway.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Return the embedded OpenTelemetry Gateway OTLP gRPC endpoint.
*/}}
{{- define "bk-nodemgr.opentelemetryGateway.otlpGrpcEndpoint" -}}
{{ include "bk-nodemgr.opentelemetryGateway.fullname" . }}:{{ dig "ports" "otlp" "servicePort" 4317 .Values.opentelemetryGateway }}
{{- end -}}

{{/*
Return the embedded Prometheus server dependency full name.
*/}}
{{- define "bk-nodemgr.prometheus.server.fullname" -}}
{{- if .Values.prometheus.server.fullnameOverride -}}
{{- .Values.prometheus.server.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default "prometheus" .Values.prometheus.nameOverride -}}
{{- $serverName := default "server" .Values.prometheus.server.name -}}
{{- if contains $name .Release.Name -}}
{{- printf "%s-%s" .Release.Name $serverName | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s-%s" .Release.Name $name $serverName | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Return the embedded Prometheus server ConfigMap name used by the official subchart override.
*/}}
{{- define "bk-nodemgr.prometheus.server.configMapName" -}}
{{- if .Values.prometheus.server.configMapOverrideName -}}
{{- printf "%s-%s" .Release.Name .Values.prometheus.server.configMapOverrideName | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- include "bk-nodemgr.prometheus.server.fullname" . -}}
{{- end -}}
{{- end -}}

{{/*
Return the embedded Loki single-binary service name.
*/}}
{{- define "bk-nodemgr.loki.fullname" -}}
{{- if .Values.loki.fullnameOverride -}}
{{- .Values.loki.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default "loki" .Values.loki.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Return the embedded Loki HTTP push/query endpoint.
*/}}
{{- define "bk-nodemgr.loki.httpEndpoint" -}}
http://{{ include "bk-nodemgr.loki.fullname" . }}.{{ .Release.Namespace }}.svc:{{ dig "loki" "server" "http_listen_port" 3100 .Values.loki }}
{{- end -}}

{{/*
Render the OpenTelemetry Gateway config. User config fully replaces defaults.
*/}}
{{- define "bk-nodemgr.opentelemetryGateway.config" -}}
{{- if .Values.opentelemetryGateway.alternateConfig -}}
{{- include "common.tplvalues.render" (dict "value" .Values.opentelemetryGateway.alternateConfig "context" $) -}}
{{- else -}}
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:{{ dig "ports" "otlp" "containerPort" 4317 .Values.opentelemetryGateway }}
      http:
        endpoint: 0.0.0.0:{{ dig "ports" "otlp-http" "containerPort" 4318 .Values.opentelemetryGateway }}
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
extensions:
  health_check:
    endpoint: 0.0.0.0:13133
service:
  extensions:
    - health_check
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
