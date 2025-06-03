package initializer

// Builder ...
type Builder interface {
	Build(*Answer) (string, error)
}
