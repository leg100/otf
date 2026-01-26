{{/*
Create the name of the cache pvc to use
*/}}
{{- define "cacheVolumeName" -}}
{{- printf "%s-cache" (include "fullname" .) }}
{{- end }}
