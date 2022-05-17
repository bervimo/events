package repositories

import (
	"testing"
	"time"

	"github.com/bervimo/events/internal/core/domain"
	"github.com/bervimo/events/utils"
	"github.com/stretchr/testify/assert"
	"github.com/tryvium-travels/memongo"
	"github.com/tryvium-travels/memongo/memongolog"
)

var mockEvent = domain.Event{
	Id:      "22",
	MatchId: "aa11",
	Code:    domain.CodeClosed,
	Head:    utils.NewBool(false),
	Next:    utils.NewString("11"),
}

var mockEvents = []domain.Event{mockEvent}

// TestInsert
func TestInsert(t *testing.T) {
	srv, _ := memongo.StartWithOptions(&memongo.Options{
		MongoVersion:   "5.0.8",
		LogLevel:       memongolog.LogLevelWarn,
		StartupTimeout: time.Duration(10) * time.Second,
	})

	opts := MongoOptions{
		URI:        srv.URI(),
		Database:   "in_memory",
		Collection: "events",
	}

	repo, err := NewMongo(opts)

	assert.NoError(t, err)
	assert.NotNil(t, repo)

	// Tested function
	res, err := repo.Insert(mockEvents)

	assert.NoError(t, err)
	assert.Equal(t, res, true)

	// Tested function
	events, err := repo.Get(mockEvent.MatchId)

	assert.NoError(t, err)

	for i, event := range events {
		assert.Equal(t, mockEvents[i].Id, event.Id)
		assert.Equal(t, mockEvents[i].MatchId, event.MatchId)
		assert.Equal(t, mockEvents[i].Code, event.Code)
		assert.Equal(t, mockEvents[i].Head, event.Head)
		assert.Equal(t, mockEvents[i].Next, event.Next)
	}
}
