package changelog

import (
	"io"

	cli "github.com/clikd-inc/cli"
)

type mockGeneratorImpl struct {
	ReturnGenerate func(io.Writer, string, *cli.Config) error
}

func (m *mockGeneratorImpl) Generate(logger *cli.Logger, w io.Writer, query string, config *cli.Config) error {
	return m.ReturnGenerate(w, query, config)
}
