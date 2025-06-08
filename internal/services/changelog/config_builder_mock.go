package changelog

type mockConfigBuilderImpl struct {
	ReturnBuild func(*Answer) (string, error)
}

func (m *mockConfigBuilderImpl) Build(ans *Answer) (string, error) {
	return m.ReturnBuild(ans)
}
