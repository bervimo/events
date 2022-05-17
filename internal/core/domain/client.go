package domain

// Client
type Client struct {
	Id       string
	MatchId  string
	Callback func(*Event, error)
	Done     <-chan struct{}
}
