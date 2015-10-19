package nntp

type Expect struct {
	Prefix string       // Prefix we expect to see
	IsErr bool          // If we got an err if so
}
