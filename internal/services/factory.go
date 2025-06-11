package services

import (
	"context"
	"fmt"

	"clikd/internal/config"
	"clikd/internal/services/ai"
	"clikd/internal/services/changelog"
	"clikd/internal/services/git"
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

// CreateGitService creates a Git service
func (f *ServiceFactory) CreateGitService() (git.Service, error) {
	f.logger.Debug("Creating Git service")
	return git.NewService()
}

// CreateGitServiceWithRepoDir creates a Git service for a specific repository directory
func (f *ServiceFactory) CreateGitServiceWithRepoDir(repoDir string) (git.Service, error) {
	f.logger.Debug("Creating Git service for repository: %s", repoDir)
	return git.NewServiceWithRepoDir(repoDir)
}

// CreateAIService creates an AI service with configuration from global config
func (f *ServiceFactory) CreateAIService() (ai.Service, error) {
	f.logger.Debug("Creating AI service with provider: %s, model: %s", f.config.AI.Provider, f.config.AI.Model)

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
	f.logger.Debug("Creating Changelog service with config: %s", configPath)

	// For now, use the existing changelog service creation
	// TODO: Refactor changelog service to accept injected dependencies
	return changelog.NewService(configPath), nil
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
		f.logger.Warn("Failed to create AI service, changelog will work without AI enhancement: %v", err)
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
