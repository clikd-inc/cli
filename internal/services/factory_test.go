package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"clikd/internal/config"
)

func TestNewServiceFactory(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (func(), error) // setup function returns cleanup function
		wantErr bool
	}{
		{
			name: "successful factory creation with valid config",
			setup: func() (func(), error) {
				// Create temporary directory for test config
				tempDir, err := os.MkdirTemp("", "clikd-test-*")
				if err != nil {
					return nil, err
				}

				// Create minimal config file
				configPath := filepath.Join(tempDir, "config.toml")
				configContent := `[general]
log_level = "info"

[ai]
provider = "openai"
model = "gpt-3.5-turbo"
api_key = "test-key"
`
				err = os.WriteFile(configPath, []byte(configContent), 0644)
				if err != nil {
					os.RemoveAll(tempDir)
					return nil, err
				}

				// Initialize config with test file
				err = config.Initialize(configPath)
				if err != nil {
					os.RemoveAll(tempDir)
					return nil, err
				}

				cleanup := func() {
					config.Reset()
					os.RemoveAll(tempDir)
				}

				return cleanup, nil
			},
			wantErr: false,
		},
		{
			name: "factory creation with default config",
			setup: func() (func(), error) {
				// Use default config behavior
				return func() {}, nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}
			defer cleanup()

			ctx := context.Background()
			factory, err := NewServiceFactory(ctx)

			if tt.wantErr {
				if err == nil {
					t.Error("NewServiceFactory() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewServiceFactory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if factory == nil {
				t.Error("NewServiceFactory() returned nil factory")
			}

			// Verify factory components
			if factory.config == nil {
				t.Error("Factory config is nil")
			}

			if factory.logger == nil {
				t.Error("Factory logger is nil")
			}

			if factory.ctx == nil {
				t.Error("Factory context is nil")
			}

			// Test getter methods
			if factory.GetConfig() == nil {
				t.Error("GetConfig() returned nil")
			}

			if factory.GetLogger() == nil {
				t.Error("GetLogger() returned nil")
			}

			if factory.GetContext() == nil {
				t.Error("GetContext() returned nil")
			}
		})
	}
}

func TestServiceFactory_CreateGitService(t *testing.T) {
	ctx := context.Background()
	factory, err := NewServiceFactory(ctx)
	if err != nil {
		t.Fatalf("Failed to create factory: %v", err)
	}

	service, err := factory.CreateGitService()
	if err != nil {
		t.Errorf("CreateGitService() error = %v", err)
		return
	}

	if service == nil {
		t.Error("CreateGitService() returned nil service")
	}
}

func TestServiceFactory_CreateGitServiceWithOptions(t *testing.T) {
	ctx := context.Background()
	factory, err := NewServiceFactory(ctx)
	if err != nil {
		t.Fatalf("Failed to create factory: %v", err)
	}

	tests := []struct {
		name             string
		repoDir          string
		tagFilterPattern string
		tagSortBy        string
		wantErr          bool
	}{
		{
			name:             "valid options",
			repoDir:          ".",
			tagFilterPattern: "v*",
			tagSortBy:        "version",
			wantErr:          false,
		},
		{
			name:             "empty options",
			repoDir:          "",
			tagFilterPattern: "",
			tagSortBy:        "",
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := factory.CreateGitServiceWithOptions(tt.repoDir, tt.tagFilterPattern, tt.tagSortBy)

			if tt.wantErr {
				if err == nil {
					t.Error("CreateGitServiceWithOptions() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CreateGitServiceWithOptions() error = %v", err)
				return
			}

			if service == nil {
				t.Error("CreateGitServiceWithOptions() returned nil service")
			}
		})
	}
}

func TestServiceFactory_CreateAIService(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (func(), error)
		wantErr bool
	}{
		{
			name: "AI service creation with missing API key should fail",
			setup: func() (func(), error) {
				tempDir, err := os.MkdirTemp("", "clikd-test-*")
				if err != nil {
					return nil, err
				}

				configPath := filepath.Join(tempDir, "config.toml")
				configContent := `[general]
log_level = "info"

[ai]
provider = "openai"
model = "gpt-3.5-turbo"
api_key = ""
api_url = "https://api.openai.com/v1"
tokens_max_input = 4000
tokens_max_output = 1000
`
				err = os.WriteFile(configPath, []byte(configContent), 0644)
				if err != nil {
					os.RemoveAll(tempDir)
					return nil, err
				}

				err = config.Initialize(configPath)
				if err != nil {
					os.RemoveAll(tempDir)
					return nil, err
				}

				cleanup := func() {
					config.Reset()
					os.RemoveAll(tempDir)
				}

				return cleanup, nil
			},
			wantErr: true,
		},
		{
			name: "AI service creation with valid API key",
			setup: func() (func(), error) {
				tempDir, err := os.MkdirTemp("", "clikd-test-*")
				if err != nil {
					return nil, err
				}

				configPath := filepath.Join(tempDir, "config.toml")
				configContent := `[general]
log_level = "info"

[ai]
provider = "openai"
model = "gpt-3.5-turbo"
api_key = "sk-test-key-with-proper-length-for-validation"
api_url = "https://api.openai.com/v1"
tokens_max_input = 4000
tokens_max_output = 1000
`
				err = os.WriteFile(configPath, []byte(configContent), 0644)
				if err != nil {
					os.RemoveAll(tempDir)
					return nil, err
				}

				err = config.Initialize(configPath)
				if err != nil {
					os.RemoveAll(tempDir)
					return nil, err
				}

				cleanup := func() {
					config.Reset()
					os.RemoveAll(tempDir)
				}

				return cleanup, nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}
			defer cleanup()

			ctx := context.Background()
			factory, err := NewServiceFactory(ctx)
			if err != nil {
				t.Fatalf("Failed to create factory: %v", err)
			}

			service, err := factory.CreateAIService()

			if tt.wantErr {
				if err == nil {
					t.Error("CreateAIService() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CreateAIService() error = %v", err)
				return
			}

			if service == nil {
				t.Error("CreateAIService() returned nil service")
			}
		})
	}
}

func TestServiceFactory_CreateChangelogService(t *testing.T) {
	ctx := context.Background()
	factory, err := NewServiceFactory(ctx)
	if err != nil {
		t.Fatalf("Failed to create factory: %v", err)
	}

	tests := []struct {
		name       string
		configPath string
		wantErr    bool
	}{
		{
			name:       "valid config path",
			configPath: "test-config.yml",
			wantErr:    false,
		},
		{
			name:       "empty config path",
			configPath: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := factory.CreateChangelogService(tt.configPath)

			if tt.wantErr {
				if err == nil {
					t.Error("CreateChangelogService() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CreateChangelogService() error = %v", err)
				return
			}

			if service == nil {
				t.Error("CreateChangelogService() returned nil service")
			}
		})
	}
}

func TestServiceFactory_CreateUpdateService(t *testing.T) {
	ctx := context.Background()
	factory, err := NewServiceFactory(ctx)
	if err != nil {
		t.Fatalf("Failed to create factory: %v", err)
	}

	service := factory.CreateUpdateService()
	if service == nil {
		t.Error("CreateUpdateService() returned nil service")
	}
}

func TestServiceFactory_CreateUpdateServiceWithOptions(t *testing.T) {
	ctx := context.Background()
	factory, err := NewServiceFactory(ctx)
	if err != nil {
		t.Fatalf("Failed to create factory: %v", err)
	}

	tests := []struct {
		name      string
		repoOwner string
		repoName  string
		timeout   time.Duration
	}{
		{
			name:      "valid options",
			repoOwner: "testowner",
			repoName:  "testrepo",
			timeout:   30 * time.Second,
		},
		{
			name:      "empty options",
			repoOwner: "",
			repoName:  "",
			timeout:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := factory.CreateUpdateServiceWithOptions(tt.repoOwner, tt.repoName, tt.timeout)
			if service == nil {
				t.Error("CreateUpdateServiceWithOptions() returned nil service")
			}
		})
	}
}

func TestServiceFactory_GetMethods(t *testing.T) {
	ctx := context.Background()
	factory, err := NewServiceFactory(ctx)
	if err != nil {
		t.Fatalf("Failed to create factory: %v", err)
	}

	// Test GetConfig
	config := factory.GetConfig()
	if config == nil {
		t.Error("GetConfig() returned nil")
	}

	// Test GetLogger
	logger := factory.GetLogger()
	if logger == nil {
		t.Error("GetLogger() returned nil")
	}

	// Test GetContext
	context := factory.GetContext()
	if context == nil {
		t.Error("GetContext() returned nil")
	}

	// Verify context is the same as provided
	if context != ctx {
		t.Error("GetContext() returned different context than provided")
	}
}
