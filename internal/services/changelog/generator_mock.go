package changelog

import (
	"io"

	"clikd/internal/utils"
)

type mockGeneratorImpl struct {
	ReturnGenerate func(io.Writer, string, *Config) error
}

func (m *mockGeneratorImpl) Generate(logger utils.Logger, w io.Writer, query string, config *Config) error {
	return m.ReturnGenerate(w, query, config)
}
