package changelog

import (
	"github.com/charmbracelet/glamour"
)

// StandardPreview contains an example of the standard style
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

// CoolPreview contains an example of the cool style
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

// KACPreview contains an example of the Keep-a-Changelog style
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

// RenderPreview renders the provided markdown with Glamour using Tokyo Night style
func RenderPreview(markdown string) (string, error) {
	return glamour.Render(markdown, "tokyo-night")
}

// GetPreview returns the preview for the specified style
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
