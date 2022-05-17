package domain

// Error type
type Error struct {
	Code uint32
	error
}

// Errors
const (
	ErrClientNotFound uint32 = 1 << iota
	ErrClientConnected
)
