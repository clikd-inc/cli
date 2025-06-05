package initializer

import (
	"os"
	"path/filepath"

	"clikd/pkg/internal/changelog"

	"github.com/BurntSushi/toml"
)

// ConfigLoader ...
type ConfigLoader interface {
	Load(*CLIContext) (*changelog.Config, error)
}

type configLoaderImpl struct {
}

// NewConfigLoader ...
func NewConfigLoader() ConfigLoader {
	return &configLoaderImpl{}
}

// Load ...
func (loader *configLoaderImpl) Load(ctx *CLIContext) (*changelog.Config, error) {
	if ctx == nil {
		return nil, changelog.ErrNotSpecifiedCLIContext
	}

	fp := filepath.Clean(ctx.ConfigPath)
	bytes, err := os.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	config := &Config{}

	// Nur TOML-Konfiguration parsen, unabhängig von der Dateierweiterung
	_, err = toml.Decode(string(bytes), config)
	if err != nil {
		return nil, err
	}

	if err := config.Normalize(ctx); err != nil {
		return nil, err
	}

	return config.Convert(ctx), nil
}
