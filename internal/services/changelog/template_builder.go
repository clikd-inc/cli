package changelog

const templateTagNameAnchor = "<a name=\"{{ .Tag.Name }}\"></a>\n"

// TemplateBuilder ...
type TemplateBuilder interface {
	Builder
}

// TemplateBuilderFactory erzeugt Builder basierend auf Template-Typ
// Wird in der aktuellen Implementierung nicht verwendet, aber für zukünftige Erweiterungen beibehalten
//
//nolint:unused
func TemplateBuilderFactory(template string) TemplateBuilder {
	switch template {
	case tplKeepAChangelog.display:
		return NewKACTemplateBuilder()
	default:
		return NewCustomTemplateBuilder()
	}
}
