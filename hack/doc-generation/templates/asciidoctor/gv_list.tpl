{{- define "gvList" -}}
{{- $groupVersions := . -}}

// Generated documentation. Please do not edit.
:page_id: api-reference
:anchor_prefix: k8s-api

[id="{p}-{page_id}"]
= API Reference

.Packages
{{- range $index, $element := $groupVersions }}
* {{ asciidocRenderGVLink . }}
{{- if $element.Kinds  }}
{{- range $element.SortedKinds }}
** {{ $element.TypeForKind . | asciidocRenderTypeLink }}
{{- end }}
{{ end }}
{{- end }}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}
