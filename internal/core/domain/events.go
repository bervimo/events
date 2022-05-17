package domain

const (
	CodeUnspecified = iota
	CodeStartOfMatch
	CodeEndOfMatch
	CodeStartOfQuarter
	CodeEndOfQuarter
	CodePostponed
	CodeClosed
)

type Code uint

type Event struct {
	Id      string  `bson:"id" json:"id"`
	MatchId string  `bson:"matchId" json:"matchId"`
	Code    Code    `bson:"code" json:"code"`
	Head    *bool   `bson:"head,omitempty" json:"head,omitempty"`
	Next    *string `bson:"next,omitempty" json:"next,omitempty"`

	CreatedAt *string `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt *string `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}
