DEBUG TEMPLATE:

Unreleased Commits Count: {{ len .Unreleased.Commits }}
Unreleased CommitGroups Count: {{ len .Unreleased.CommitGroups }}

{{ if .Unreleased.CommitGroups -}}
UNRELEASED COMMIT GROUPS:
{{ range .Unreleased.CommitGroups -}}
- Group: {{ .Title }} ({{ len .Commits }} commits)
{{ end -}}
{{ else -}}
NO UNRELEASED COMMIT GROUPS
{{ end -}}

Versions Count: {{ len .Versions }}
{{ range .Versions }}
Version: {{ .Tag.Name }} ({{ len .Commits }} commits, {{ len .CommitGroups }} groups)
{{ end }} 
