package services

import (
	"sync"
	"testing"
	"time"

	"github.com/bervimo/events/internal/core/domain"
	"github.com/bervimo/events/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var mockEvent = domain.Event{
	Id:      "22",
	MatchId: "aa11",
	Code:    domain.CodeClosed,
	Head:    utils.NewBool(false),
	Next:    utils.NewString("11"),
}

var mockEvents = []domain.Event{mockEvent}

var mockedClient = domain.Client{
	Id:       "00",
	MatchId:  mockEvent.MatchId,
	Callback: func(e *domain.Event, err error) {},
	Done:     make(<-chan struct{}),
}

type mockedRepository struct {
	mock.Mock
	wg sync.WaitGroup
}

func (mr *mockedRepository) Insert(events []domain.Event) (bool, error) {
	args := mr.Called(events)

	return args.Bool(0), args.Error(1)
}

func (mr *mockedRepository) Get(matchId string) ([]domain.Event, error) {
	args := mr.Called(matchId)

	return args.Get(0).([]domain.Event), args.Error(1)
}

func (mr *mockedRepository) OnEvents(callback func(event *domain.Event, err error)) {
	mr.Called(callback)

	mr.wg.Done()
}

func (mr *mockedRepository) Close() error {
	args := mr.Called()

	return args.Error(0)
}

// TestNewService
func TestNewService(t *testing.T) {
	mc := mock.AnythingOfType("func(*domain.Event, error)")
	mr := mockedRepository{}

	mr.On("OnEvents", mc).Return(nil)

	go (func() {
		mr.wg.Add(1)
		mr.wg.Wait()

		mr.AssertExpectations(t)
	})()

	// Tested function
	srv := NewService(&mr)

	assert.NotEmpty(t, srv)
}

// TestInsert
func TestInsert(t *testing.T) {
	mr := mockedRepository{}

	mr.On("Insert", mockEvents).Return(true, nil)

	srv := Service{repository: &mr}

	// Tested function
	res, err := srv.Insert(mockEvents)

	assert.Equal(t, res, true)
	assert.Equal(t, err, nil)

	mr.AssertExpectations(t)
}

// TestGet
func TestGet(t *testing.T) {
	mr := mockedRepository{}

	ms := Service{repository: &mr}

	mr.On("Get", mockEvent.MatchId).Return(mockEvents, nil)

	// Tested function
	res, err := ms.Get(mockEvent.MatchId)

	assert.Equal(t, res, mockEvents)
	assert.Equal(t, err, nil)

	mr.AssertExpectations(t)
}

// TestOnEvents
func TestOnEvents(t *testing.T) {
	clients := map[string]domain.Client{}

	srv := Service{clients: clients}

	// Tested function
	go (func() {
		err := srv.OnEvents(mockedClient)

		assert.Equal(t, err, nil)
	})()

	time.AfterFunc(time.Millisecond*10, func() { <-mockedClient.Done })
}
