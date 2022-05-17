package repositories

import (
	"testing"

	"github.com/bervimo/events/internal/core/domain"
	"github.com/bervimo/events/utils"
	"github.com/stretchr/testify/assert"
	"github.com/tryvium-travels/memongo"
	"github.com/tryvium-travels/memongo/memongolog"
)

// FuzzInsert
func FuzzInsert(f *testing.F) {
	srv, err := memongo.StartWithOptions(&memongo.Options{
		MongoVersion: "5.0.8",
		LogLevel:     memongolog.LogLevelWarn,
	})

	assert.NoError(f, err)
	assert.NotNil(f, srv)

	defer srv.Stop()

	opts := MongoOptions{
		URI:        srv.URI(),
		Database:   "in_memory",
		Collection: "events",
	}

	repo, err := NewMongo(opts)

	assert.NoError(f, err)
	assert.NotNil(f, repo)

	seed := [][]any{
		{"foo", "foo", uint(1), true, "foo"},
		{"", "", uint(2), false, ""},
		{"!@acb123?", "!@acb123?", uint(3), false, "!@acb123?"},
	}

	for _, args := range seed {
		f.Add(
			args[0].(string), // Id
			args[1].(string), // MatchId
			args[2].(uint),   // Code
			args[3].(bool),   // Head
			args[4].(string), // Next
		)
	}

	f.Fuzz(func(t *testing.T, a string, b string, c uint, d bool, e string) {
		me := []domain.Event{
			{
				Id:      a,
				MatchId: b,
				Code:    domain.Code(c),
				Head:    utils.NewBool(d),
				Next:    utils.NewString(e),
			},
		}

		_, err := repo.Insert(me)

		assert.NoError(t, err)
	})
}
