package changelog

// Builder ...
type Builder interface {
	Build(*Answer) (string, error)
}
