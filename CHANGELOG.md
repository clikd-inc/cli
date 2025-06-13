<a name="unreleased"></a>
## [Unreleased]

### Code Refactoring
- **ai:** Here are the processed commits following your strict requirements:
- remove unused methods from Service interface to simplify code and improve maintainability
- **config:** chore(ai): delete corresponding tests for removed methods to keep test suite clean

### Features
- **ai:** refactor(usecases): remove unused types and functions related to commit categorization and enhancement to streamline codebase
- **changelog:** replace ConfigData with Config to simplify configuration structure and improve clarity
- **changelog:** refactor(config): remove unnecessary conversion functions and streamline configuration handling
- **changelog:** refactor(config): update configuration loading and saving methods to use new Config structure
- **changelog:** refactor(config): enhance environment variable handling for configuration settings
- **main.go:** refactor(validation): improve validation logic for provider and model compatibility
- **services:** refactor AI client and service initialization to use global configuration parameters for improved flexibility and maintainability
- **update:** refactor(ai): remove deprecated Config struct and replace with CreateModelConfig function for better clarity and structure
- **update:** chore(ai): update README to reflect changes in configuration loading and service creation process


<a name="v0.1.0"></a>
## v0.1.0 - 2025-06-09
### Features
- integrate AI service for changelog enhancement to improve readability and professionalism
- **ai:** refactor(changelog): update changelog generation to support AI enhancements and dependency injection
- **changelog:** fix(ai): adjust AI model configuration to use dynamic parameters from config
- **changelog:** chore(ai): remove deprecated methods and improve AI service integration
- **changelog:** style(changelog): enhance comments and code structure for better clarity and maintainability
- **changelog:** implement a new changelog generation service with improved options and UI support
- **changelog:** chore(changelog): add a new binary file clikd-test for testing purposes
- **changelog:** refactor(changelog): clean up and modularize changelog generation logic for better maintainability
- **changelog:** add a new changelog file to document project changes and updates for better tracking and transparency
- **changelog:** chore(docs): remove outdated changelog integration plan document to streamline documentation and avoid confusion
- **config:** fix(config): remove AI enable flag from configuration as AI is now always enabled for consistency and simplicity
- **initialize:** feat(ai): enhance AI integration by allowing advanced configuration options for token limits and custom API endpoints
- **test:** refactor(ai): simplify API key retrieval logic to use a single environment variable for all AI providers
- **update:** style(ui): improve user interface for AI configuration steps to enhance user experience during setup


[Unreleased]: https://github.com/clikd-inc/cli.git/compare/v0.1.0...HEAD
