# CHANGELOG Jira Integration

{{ range .Versions }}
<a name="{{ .Tag.Name }}"></a>
## {{ .Tag.Name }} - {{ datetime "2006-01-02" .Tag.Date }}
{{ range .CommitGroups }}
### {{ .Title }}
{{ range .Commits }}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}{{ if .JiraIssue }}
  **Jira:** [{{ .JiraIssue.Key }}]({{ $.Info.RepositoryURL }}/browse/{{ .JiraIssue.Key }})
  **Summary:** {{ .JiraIssue.Summary }}
  **Type:** {{ .JiraIssue.Type }}
  {{ if .JiraIssue.Labels }}**Labels:** {{ join ", " .JiraIssue.Labels }}{{ end }}
  {{ if .JiraIssue.Description }}**Description:** {{ indent (replace .JiraIssue.Description "\n" " " -1) 4 }}{{ end }}
{{ end }}
{{ end }}
{{ end }}

{{- if .RevertCommits -}}
### Reverts
{{ range .RevertCommits }}
- {{ .Revert.Header }}
{{- end }}
{{ end -}}

{{- if .MergeCommits -}}
### Pull Requests
{{ range .MergeCommits }}
- {{ .Header }}
{{- end }}
{{ end -}}

{{- if .NoteGroups -}}
{{ range .NoteGroups }}
### {{ .Title }}
{{ range .Notes }}
{{ .Body }}
{{ end }}
{{ end -}}
{{ end -}}
{{ end -}}

{{- if .Versions }}
[Unreleased]: {{ .Info.RepositoryURL }}/compare/{{ $latest := index .Versions 0 }}{{ $latest.Tag.Name }}...HEAD
{{ range .Versions -}}
{{ if .Tag.Previous -}}
[{{ .Tag.Name }}]: {{ $.Info.RepositoryURL }}/compare/{{ .Tag.Previous.Name }}...{{ .Tag.Name }}
{{ end -}}
{{ end -}}
{{ end -}} 
