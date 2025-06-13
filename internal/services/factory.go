package services

import (
	"context"
	"fmt"
	"time"

	"clikd/internal/config"
	"clikd/internal/services/ai"
	"clikd/internal/services/changelog"
	"clikd/internal/services/git"
	"clikd/internal/services/update"
	"clikd/internal/utils"
)

// ServiceFactory manages the creation of all services with proper dependency injection
type ServiceFactory struct {
	config *config.Config
	logger utils.Logger
	ctx    context.Context
}

// NewServiceFactory creates a new service factory
func NewServiceFactory(ctx context.Context) (*ServiceFactory, error) {
	// Load global configuration
	cfg, err := config.EnsureInitialized()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create logger from config
	logger := utils.NewLogger(cfg.General.LogLevel, true)

	return &ServiceFactory{
		config: cfg,
		logger: logger,
		ctx:    ctx,
	}, nil
}

// CreateGitService creates a Git service with all dependencies
func (f *ServiceFactory) CreateGitService() (git.Service, error) {
	f.logger.Debug("Creating Git service")

	// Create git service with injected logger
	service, err := git.NewService()
	if err != nil {
		return nil, fmt.Errorf("failed to create git service: %w", err)
	}

	return service, nil
}

// CreateGitServiceWithOptions creates a Git service with custom options
func (f *ServiceFactory) CreateGitServiceWithOptions(repoDir, tagFilterPattern, tagSortBy string) (git.Service, error) {
	f.logger.Debug("Creating Git service with options", "repoDir", repoDir, "tagFilter", tagFilterPattern, "tagSort", tagSortBy)

	// Create service with options and injected logger
	service, err := git.NewServiceWithOptions(repoDir, tagFilterPattern, tagSortBy, f.logger.WithFields(map[string]interface{}{"module": "git"}))
	if err != nil {
		return nil, fmt.Errorf("failed to create git service with options: %w", err)
	}

	return service, nil
}

// CreateAIService creates an AI service with configuration from global config
func (f *ServiceFactory) CreateAIService() (ai.Service, error) {
	f.logger.Debug("Creating AI service", "provider", f.config.AI.Provider, "model", f.config.AI.Model)

	return ai.NewService(
		f.ctx,
		f.config.AI.Provider,
		f.config.AI.Model,
		f.config.AI.APIKey,
		f.config.AI.APIURL,
		f.config.AI.TokensMaxInput,
		f.config.AI.TokensMaxOutput,
	)
}

// CreateChangelogService creates a Changelog service with all dependencies
func (f *ServiceFactory) CreateChangelogService(configPath string) (*changelog.Service, error) {
	f.logger.Debug("Creating Changelog service", "config", configPath)

	// Create the service with the new high-level functionality
	return changelog.NewService(configPath), nil
}

// CreateChangelogServiceWithOptions creates a Changelog service and prepares generation options
func (f *ServiceFactory) CreateChangelogServiceWithOptions(configPath string, options *changelog.GenerationOptions) (*changelog.Service, *changelog.GenerationResult, error) {
	f.logger.Debug("Creating Changelog service with options")

	// Create service
	service := changelog.NewService(configPath)

	// Prepare generation with the provided options
	result, err := service.PrepareGeneration(f.ctx, options)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare changelog generation: %w", err)
	}

	return service, result, nil
}

// CreateChangelogServiceWithDependencies creates a Changelog service with injected dependencies
// This is the future-ready version that will be used once we refactor the changelog service
func (f *ServiceFactory) CreateChangelogServiceWithDependencies(configPath string) (*changelog.Service, error) {
	f.logger.Debug("Creating Changelog service with injected dependencies")

	// Create Git service
	gitService, err := f.CreateGitService()
	if err != nil {
		return nil, fmt.Errorf("failed to create Git service: %w", err)
	}

	// Create AI service
	aiService, err := f.CreateAIService()
	if err != nil {
		f.logger.Warn("Failed to create AI service, changelog will work without AI enhancement", "error", err)
		// AI service is optional for changelog, so we continue without it
		aiService = nil
	}

	// TODO: Once changelog service is refactored, use this:
	// return changelog.NewServiceWithDependencies(configPath, gitService, aiService), nil

	// For now, fall back to the existing service
	_ = gitService // Suppress unused variable warning
	_ = aiService  // Suppress unused variable warning
	return changelog.NewService(configPath), nil
}

// CreateChangelogServiceWithAI creates a Changelog service with AI enhancement
// This method properly injects the AI service into the changelog generator
func (f *ServiceFactory) CreateChangelogServiceWithAI(configPath string) (*changelog.Service, error) {
	f.logger.Debug("Creating Changelog service with AI enhancement")

	// Create service with factory injection for AI enhancement
	service := changelog.NewServiceWithFactory(configPath, f)

	// Verify AI service is available
	_, err := f.CreateAIService()
	if err != nil {
		f.logger.Debug("Could not create AI service, changelog will work without AI enhancement: %v", err)
		// Return service anyway - it will gracefully degrade
		return service, nil
	}

	f.logger.Debug("AI service is available for changelog enhancement")
	return service, nil
}

// CreateChangelogGeneratorWithAI creates a changelog generator with AI service injected
// This is a helper method for proper AI service injection
func (f *ServiceFactory) CreateChangelogGeneratorWithAI(config *changelog.Config, logger utils.Logger) (*changelog.Generator, error) {
	f.logger.Debug("Creating Changelog generator with AI enhancement")

	// Create the basic generator
	generator := changelog.NewGenerator(logger, config)

	// Try to create and inject AI service
	aiService, err := f.CreateAIService()
	if err != nil {
		f.logger.Debug("Could not create AI service, generator will work without AI enhancement: %v", err)
		// Return generator without AI - graceful degradation
		return generator, nil
	}

	f.logger.Debug("AI service created successfully, injecting into changelog generator")
	generator.SetAIService(aiService)

	return generator, nil
}

// CreateUpdateService creates an Update service with all dependencies
func (f *ServiceFactory) CreateUpdateService() update.Service {
	f.logger.Debug("Creating Update service")

	// Create update service with injected logger
	return update.NewServiceWithOptions(nil, f.logger.WithFields(map[string]interface{}{"module": "update"}))
}

// CreateUpdateServiceWithOptions creates an Update service with custom options
func (f *ServiceFactory) CreateUpdateServiceWithOptions(repoOwner, repoName string, timeout time.Duration) update.Service {
	f.logger.Debug("Creating Update service with options", "owner", repoOwner, "repo", repoName, "timeout", timeout)

	options := &update.UpdateOptions{
		RepoOwner: repoOwner,
		RepoName:  repoName,
		Timeout:   timeout,
	}

	return update.NewServiceWithOptions(options, f.logger.WithFields(map[string]interface{}{"module": "update"}))
}

// GetConfig returns the loaded configuration
func (f *ServiceFactory) GetConfig() *config.Config {
	return f.config
}

// GetLogger returns the logger instance
func (f *ServiceFactory) GetLogger() utils.Logger {
	return f.logger
}

// GetContext returns the context
func (f *ServiceFactory) GetContext() context.Context {
	return f.ctx
}

// Implement ServiceFactoryInterface for changelog service injection
// This avoids circular import issues by using interfaces

// CreateAIService implements ServiceFactoryInterface
func (f *ServiceFactory) CreateAIServiceForChangelog() (changelog.AIServiceInterface, error) {
	aiService, err := f.CreateAIService()
	if err != nil {
		return nil, err
	}

	// Return an adapter that implements the changelog.AIServiceInterface
	return &aiServiceAdapter{service: aiService}, nil
}

// GetConfigForChangelog implements ServiceFactoryInterface
func (f *ServiceFactory) GetConfigForChangelog() changelog.ConfigInterface {
	return &configAdapter{config: f.config}
}

// aiServiceAdapter adapts ai.Service to changelog.AIServiceInterface
type aiServiceAdapter struct {
	service ai.Service
}

func (a *aiServiceAdapter) EnhanceChangelog(changelog string) (string, error) {
	return a.service.EnhanceChangelog(changelog)
}

// configAdapter adapts config.Config to changelog.ConfigInterface
type configAdapter struct {
	config *config.Config
}

func (c *configAdapter) GetAIConfig() (provider, model, apiKey, apiURL string, tokensMaxInput, tokensMaxOutput int) {
	return c.config.AI.Provider, c.config.AI.Model, c.config.AI.APIKey, c.config.AI.APIURL, c.config.AI.TokensMaxInput, c.config.AI.TokensMaxOutput
}
