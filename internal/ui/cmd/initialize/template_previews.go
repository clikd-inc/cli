package initialize

// StandardTemplatePreview contains an example of the standard style
const StandardTemplatePreview = `# CHANGELOG


<a name="1.0.0"></a>
## [1.0.0](https://github.com/clikd-inc/cli/compare/0.1.1...1.0.0) (2018-05-06)

### Bug Fixes

* **core:** Fix CHANGELOG format

### Code Refactoring

* **logger:** remove unnecessary checks

### Features

* **core:** A big change!!
* **core:** webpack 4 support

### BREAKING CHANGE


This is a example of breaking changes.

You can describe the contents over several lines, Thanks.


<a name="0.1.1"></a>
## [0.1.1](https://github.com/clikd-inc/cli/compare/0.1.0...0.1.1) (2018-05-06)

### Bug Fixes

* **core:** redirect stderr to fix mingw build
* **installer:** Fix CLI installation on Windows

### Code Refactoring

* **core:** update comment

### Features

* **core:** Add events method to Asciicast class


<a name="0.1.0"></a>
## 0.1.0 (2018-05-06)

### Bug Fixes

* **config:** Fix color configuration with an array

### Features

* **core:** Add a build step to hoist warning conditions
* **core:** First implement
* **lang:** Fixes language in error message.
`

// CoolTemplatePreview contains an example of the cool style
const CoolTemplatePreview = `# CHANGELOG

<a name="1.0.0"></a>
## [1.0.0](https://github.com/clikd-inc/cli/compare/0.1.1...1.0.0)

> 2018-05-06

### Bug Fixes

* **core:** Fix CHANGELOG format

### Code Refactoring

* **logger:** remove unnecessary checks

### Features

* **core:** A big change!!
* **core:** webpack 4 support

### BREAKING CHANGE


This is a example of breaking changes.

You can describe the contents over several lines, Thanks.


<a name="0.1.1"></a>
## [0.1.1](https://github.com/clikd-inc/cli/compare/0.1.0...0.1.1)

> 2018-05-06

### Bug Fixes

* **core:** redirect stderr to fix mingw build
* **installer:** Fix CLI installation on Windows

### Code Refactoring

* **core:** update comment

### Features

* **core:** Add events method to Asciicast class


<a name="0.1.0"></a>
## 0.1.0

> 2018-05-06

### Bug Fixes

* **config:** Fix color configuration with an array

### Features

* **core:** Add a build step to hoist warning conditions
* **core:** First implement
* **lang:** Fixes language in error message.
`

// KACTemplatePreview contains an example of the Keep-a-Changelog style
const KACTemplatePreview = `# CHANGELOG

This CHANGELOG is a format conforming to [keep-a-changelog](https://github.com/olivierlacan/keep-a-changelog).  


<a name="1.0.0"></a>
## [1.0.0] - 2018-05-06
### Bug Fixes
- **core:** Fix CHANGELOG format

### Code Refactoring
- **logger:** remove unnecessary checks

### Features
- **core:** A big change!!
- **core:** webpack 4 support

### BREAKING CHANGE

This is a example of breaking changes.

You can describe the contents over several lines, Thanks.


<a name="0.1.1"></a>
## [0.1.1] - 2018-05-06
### Bug Fixes
- **core:** redirect stderr to fix mingw build
- **installer:** Fix CLI installation on Windows

### Code Refactoring
- **core:** update comment

### Features
- **core:** Add events method to Asciicast class


<a name="0.1.0"></a>
## 0.1.0 - 2018-05-06
### Bug Fixes
- **config:** Fix color configuration with an array

### Features
- **core:** Add a build step to hoist warning conditions
- **core:** First implement
- **lang:** Fixes language in error message.


[1.0.0]: https://github.com/clikd-inc/cli/compare/0.1.1...1.0.0
[0.1.1]: https://github.com/clikd-inc/cli/compare/0.1.0...0.1.1
`

// GetTemplatePreview returns the preview for the specified template style
func GetTemplatePreview(style string) string {
	switch style {
	case "cool":
		return CoolTemplatePreview
	case "keep-a-changelog":
		return KACTemplatePreview
	default:
		return StandardTemplatePreview
	}
}
