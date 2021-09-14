{{/*
Define dynatrace-service environment variables here
*/}}
{{- define "dynatrace-service.environment-variables" -}}
- name: DATASTORE
  value: 'http://mongodb-datastore:8080'
- name: CONFIGURATION_SERVICE
  value: 'http://configuration-service:8080'
- name: SHIPYARD_CONTROLLER
  value: 'http://shipyard-controller:8080'
- name: PLATFORM
  value: kubernetes
- name: POD_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: GENERATE_TAGGING_RULES
  value: '{{ .Values.dynatraceService.config.generateTaggingRules }}'
- name: GENERATE_PROBLEM_NOTIFICATIONS
  value: '{{ .Values.dynatraceService.config.generateProblemNotifications }}'
- name: GENERATE_MANAGEMENT_ZONES
  value: '{{ .Values.dynatraceService.config.generateManagementZones }}'
- name: GENERATE_DASHBOARDS
  value: '{{ .Values.dynatraceService.config.generateDashboards }}'
- name: GENERATE_METRIC_EVENTS
  value: '{{ .Values.dynatraceService.config.generateMetricEvents }}'
- name: SYNCHRONIZE_DYNATRACE_SERVICES
  value: '{{ .Values.dynatraceService.config.synchronizeDynatraceServices }}'
- name: SYNCHRONIZE_DYNATRACE_SERVICES_INTERVAL_SECONDS
  value: '{{ .Values.dynatraceService.config.synchronizeDynatraceServicesIntervalSeconds }}'
- name: HTTP_SSL_VERIFY
  value: '{{ .Values.dynatraceService.config.httpSSLVerify }}'
- name: HTTP_PROXY
  value: '{{ .Values.dynatraceService.config.httpProxy }}'
- name: HTTPS_PROXY
  value: '{{ .Values.dynatraceService.config.httpsProxy }}'
- name: NO_PROXY
  value: '{{ .Values.dynatraceService.config.noProxy }}'
- name: LOG_LEVEL_DYNATRACE_SERVICE
  value: '{{ .Values.dynatraceService.config.logLevel }}'
- name: KEPTN_API_URL
  value: '{{ .Values.dynatraceService.config.keptnApiUrl }}'
- name: KEPTN_BRIDGE_URL
  value: '{{ .Values.dynatraceService.config.keptnBridgeUrl }}'
- name: KEPTN_API_TOKEN
  valueFrom:
    secretKeyRef:
      name: keptn-api-token
      key: keptn-api-token
{{- end }}