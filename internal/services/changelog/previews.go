package changelog

import (
	"github.com/charmbracelet/glamour"
)

// StandardPreview enthält ein Beispiel des Standard-Stils
const StandardPreview = `# Changelog

## v1.0.0 (2023-01-15)

### Features

* **api:** Add new endpoint for user profiles
* Implement authentication flow

### Bug Fixes

* Fix race condition in concurrent requests
* **ui:** Correct alignment of buttons on mobile view

### Performance Improvements

* Optimize database queries
`

// CoolPreview enthält ein Beispiel des Cool-Stils
const CoolPreview = `# Changelog

## v1.0.0

> 2023-01-15

### Features

* **api:** Add new endpoint for user profiles
* Implement authentication flow

### Bug Fixes

* Fix race condition in concurrent requests
* **ui:** Correct alignment of buttons on mobile view

### Performance Improvements

* Optimize database queries
`

// KACPreview enthält ein Beispiel des Keep-a-Changelog-Stils
const KACPreview = `# Changelog

## [Unreleased]

### Added
- New visual identity
- Version navigation

## [1.0.0] - 2023-01-15

### Added
- New authentication API
- Documentation improvements

### Changed
- Start using semantic versioning

### Fixed
- Fix race condition in concurrent requests

[Unreleased]: https://github.com/username/repo/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/username/repo/releases/tag/v1.0.0
`

// RenderPreview rendert das angegebene Markdown mit Glamour
func RenderPreview(markdown string) (string, error) {
	return glamour.Render(markdown, "dark")
}

// GetPreview gibt die Vorschau für den angegebenen Stil zurück
func GetPreview(style string) string {
	switch style {
	case "cool":
		return CoolPreview
	case "keep-a-changelog":
		return KACPreview
	default:
		return StandardPreview
	}
}
