{{/*
Render the data section of the backend config (backend_conf.yaml) with the given workflow config.
Params: dict "root" . "workflow" <workflow config values>
*/}}
{{- define "bk-nodemgr.backendConfigData" }}
    runMode: {{ .root.Values.backend.config.runMode }}

    tenantMode: {{ .root.Values.backend.config.tenantMode }}

    kms:
      enabled: {{ .root.Values.kms.enabled }}
      {{- if .root.Values.kms.enabled }}
      globalSecretFile: {{ .root.Values.common.secretPath }}/kms-global-envelope
      moduleSecretFile: {{ .root.Values.common.secretPath }}/kms-backend-envelope
      {{- end }}

    system: {{ .root.Values.backend.config.system | toYaml | nindent 6 }}

    wiki: {{ .root.Values.backend.config.wiki | toYaml | nindent 6 }}

    cryptoType: {{ .root.Values.backend.config.cryptoType | default "CLASSIC" | quote }}

    encryptKey: {{ .root.Values.backend.config.encryptKey | quote }}

    infoServer: {{ include "bk-nodemgr.httpServerConfig" (dict "config" .root.Values.backend.config.infoServer "terminationGracePeriodSeconds" (.root.Values.backend.terminationGracePeriodSeconds | default .root.Values.terminationGracePeriodSeconds)) | nindent 6 }}
      advertiseIPV4: __ADVERTISE_IPV4__
      advertiseIPV6: __ADVERTISE_IPV6__

    adminServer: {{ include "bk-nodemgr.httpServerConfig" (dict "config" .root.Values.backend.config.adminServer "terminationGracePeriodSeconds" (.root.Values.backend.terminationGracePeriodSeconds | default .root.Values.terminationGracePeriodSeconds)) | nindent 6 }}
      advertiseIPV4: __ADVERTISE_IPV4__
      advertiseIPV6: __ADVERTISE_IPV6__

    basicServer: {{ include "bk-nodemgr.httpServerConfig" (dict "config" .root.Values.backend.config.basicServer "terminationGracePeriodSeconds" (.root.Values.backend.terminationGracePeriodSeconds | default .root.Values.terminationGracePeriodSeconds)) | nindent 6 }}
      advertiseIPV4: __ADVERTISE_IPV4__
      advertiseIPV6: __ADVERTISE_IPV6__

    callbackServer: {{ include "bk-nodemgr.httpServerConfig" (dict "config" .root.Values.backend.config.callbackServer "terminationGracePeriodSeconds" (.root.Values.backend.terminationGracePeriodSeconds | default .root.Values.terminationGracePeriodSeconds)) | nindent 6 }}
      advertiseIPV4: __ADVERTISE_IPV4__
      advertiseIPV6: __ADVERTISE_IPV6__

    proxyServer: {{ include "bk-nodemgr.httpServerConfig" (dict "config" .root.Values.backend.config.proxyServer "terminationGracePeriodSeconds" (.root.Values.backend.terminationGracePeriodSeconds | default .root.Values.terminationGracePeriodSeconds)) | nindent 6 }}
      advertiseIPV4: __ADVERTISE_IPV4__
      advertiseIPV6: __ADVERTISE_IPV6__

    workflow: {{ include "bk-nodemgr.workflowConfig" (dict "config" .workflow "terminationGracePeriodSeconds" (.root.Values.backend.terminationGracePeriodSeconds | default .root.Values.terminationGracePeriodSeconds)) | nindent 6 }}

    cmdb: {{ .root.Values.backend.config.cmdb | toYaml | nindent 6 }}

    gse: {{ .root.Values.backend.config.gse | toYaml | nindent 6 }}

    userManager: {{ .root.Values.backend.config.userManager | toYaml | nindent 6 }}

    monitor: {{ .root.Values.backend.config.monitor | toYaml | nindent 6 }}

    nodeEventDataID: {{ .root.Values.backend.config.nodeEventDataID | toYaml | nindent 6 }}

    creditVault: {{ .root.Values.backend.config.creditVault | toYaml | nindent 6 }}

    log:
      {{- include "bk-nodemgr.logConfig" (dict "config" .root.Values.backend.config.log "context" .root) | nindent 6 }}

    iamV3: {{ .root.Values.backend.config.iamV3 | toYaml | nindent 6 }}

    iamV4: {{ .root.Values.backend.config.iamV4 | toYaml | nindent 6 }}

    file: {{ .root.Values.backend.config.file | toYaml | nindent 6 }}

    etcd:
      {{- if .root.Values.etcd.enabled }}
      endpoints:
        - {{ include "bk-nodemgr.etcd.fullname" .root }}-headless:{{ include "bk-nodemgr.etcd.clientPort" .root }}
      {{- else }}
      endpoints: {{ .root.Values.externalEtcd.endpoints | toYaml | nindent 6 }}
      username: {{ .root.Values.externalEtcd.username }}
      password: {{ .root.Values.externalEtcd.password }}
      tls:
        {{- if len .root.Values.externalEtcd.tls.ca }}
        caFile: {{ .root.Values.common.certPath }}/etcdca.crt
        certFile: {{ .root.Values.common.certPath }}/etcd.crt
        keyFile: {{ .root.Values.common.certPath }}/etcd.key
        password: {{ .root.Values.externalEtcd.tls.password }}
        {{- else if len .root.Values.externalEtcd.tls.secretName }}
        caFile: {{ .root.Values.common.certPath }}/{{ .root.Values.externalEtcd.tls.secretCaFileName }}
        certFile: {{ .root.Values.common.certPath }}/{{ .root.Values.externalEtcd.tls.secretCertFileName }}
        keyFile: {{ .root.Values.common.certPath }}/{{ .root.Values.externalEtcd.tls.secretKeyFileName }}
        {{- end}}
      {{- end}}

    redis:
      {{- if .root.Values.redis.enabled }}
      type: standalone
      addrs:
        - "{{ template "bk-nodemgr.fullname" .root }}-redis-headless:{{ .root.Values.redis.service.ports.redis }}"
      password: {{ .root.Values.redis.auth.password }}
      db: 0
      {{- else }}
      type: {{ .root.Values.externalRedis.type }}
      addrs: {{ .root.Values.externalRedis.addrs | toYaml | nindent 8 }}
      password: {{ .root.Values.externalRedis.password }}
      db: {{ .root.Values.externalRedis.db }}
      {{- if eq .root.Values.externalRedis.type "sentinel" }}
      masterName: {{ .root.Values.externalRedis.masterName }}
      {{- end }}
      tls:
        {{- if len .root.Values.externalRedis.tls.ca }}
        caFile: {{ .root.Values.common.certPath }}/redisca.crt
        certFile: {{ .root.Values.common.certPath }}/redis.crt
        keyFile: {{ .root.Values.common.certPath }}/redis.key
        password: {{ .root.Values.externalRedis.tls.password }}
        {{- else if len .root.Values.externalRedis.tls.secretName }}
        caFile: {{ .root.Values.common.certPath }}/{{ .root.Values.externalRedis.tls.secretCaFileName }}
        certFile: {{ .root.Values.common.certPath }}/{{ .root.Values.externalRedis.tls.secretCertFileName }}
        keyFile: {{ .root.Values.common.certPath }}/{{ .root.Values.externalRedis.tls.secretKeyFileName }}
        {{- end}}
      {{- end}}
      traceServiceName: {{ .root.Values.backend.config.redis.traceServiceName }}
      traceSampleRate: {{ .root.Values.backend.config.redis.traceSampleRate }}

    mongodb:
      appName: {{ template "bk-nodemgr.fullname" .root }}-backend
      {{- if .root.Values.mongodb.enabled }}
      replicaSet: {{ .root.Values.mongodb.replicaSetName }}
      hosts:
        - {{ template "bk-nodemgr.fullname" .root }}-mongodb-headless
      username: {{ .root.Values.mongodb.auth.username }}
      password: {{ .root.Values.mongodb.auth.password }}
      database: {{ .root.Values.mongodb.auth.database }}
      authSource: {{ .root.Values.mongodb.auth.database }}
      authMechanism: SCRAM-SHA-256
      {{- else }}
      replicaSet: {{ .root.Values.externalMongodb.replicaSetName }}
      hosts: {{ .root.Values.externalMongodb.hosts | toYaml | nindent 6}}
      username: {{ .root.Values.externalMongodb.username }}
      password: {{ .root.Values.externalMongodb.password }}
      database: {{ .root.Values.externalMongodb.database }}
      authSource: {{ .root.Values.externalMongodb.authSource }}
      authMechanism: {{ .root.Values.externalMongodb.authMechanism }}
      tls:
        {{- if len .root.Values.externalMongodb.tls.ca }}
        caFile: {{ .root.Values.common.certPath }}/mongoca.crt
        certFile: {{ .root.Values.common.certPath }}/mongo.crt
        keyFile: {{ .root.Values.common.certPath }}/mongo.key
        password: {{ .root.Values.externalMongodb.tls.password }}
        {{- else if len .root.Values.externalMongodb.tls.secretName }}
        caFile: {{ .root.Values.common.certPath }}/{{ .root.Values.externalMongodb.tls.secretCaFileName }}
        certFile: {{ .root.Values.common.certPath }}/{{ .root.Values.externalMongodb.tls.secretCertFileName }}
        keyFile: {{ .root.Values.common.certPath }}/{{ .root.Values.externalMongodb.tls.secretKeyFileName }}
        {{- end}}
      {{- end}}
      traceServiceName: {{ .root.Values.backend.config.mongodb.traceServiceName }}
      traceSampleRate: {{ .root.Values.backend.config.mongodb.traceSampleRate }}

    installerFileGroup: {{ .root.Values.backend.config.installerFileGroup | toYaml | nindent 6 }}

    networkUnit: {{ .root.Values.backend.config.networkUnit | toYaml | nindent 6 }}

    fileCache: {{ .root.Values.backend.config.fileCache | toYaml | nindent 6 }}

    gseDeployConfs: {{ .root.Values.backend.config.gseDeployConfs | toYaml | nindent 6 }}

    access:
      {{- if hasKey (.root.Values.backend.config.access | default dict) "virtualUser" }}
      virtualUser: {{ .root.Values.backend.config.access.virtualUser | quote }}
      {{- else }}
      virtualUser: {{ ternary "admin" "bk_admin" (eq .root.Values.backend.config.tenantMode "single") | quote }}
      {{- end }}

    profiling:
      {{- include "bk-nodemgr.profilingConfig" (dict "config" .root.Values.backend.config.profiling "context" .root) | nindent 6 }}

    tracing:
      {{- include "bk-nodemgr.tracingConfig" (dict "config" .root.Values.backend.config.tracing "context" .root) | nindent 6 }}
      instanceID: "__TRACING_INSTANCEID__"
{{- end -}}
