package services

import (
	"sync"

	"github.com/bervimo/events/internal/core/domain"
	"github.com/bervimo/events/internal/core/ports"
)

type Service struct {
	repository ports.Repository

	clients map[string]domain.Client
	mutex   sync.Mutex
}

// NewService
func NewService(repository ports.Repository) *Service {
	srv := &Service{
		repository: repository,
		clients:    map[string]domain.Client{},
	}

	go (func() {
		srv.repository.OnEvents(func(event *domain.Event, err error) {
			for _, cli := range srv.clients {
				cli.Callback(event, err)
			}
		})
	})()

	return srv
}

// Insert
func (srv *Service) Insert(events []domain.Event) (bool, error) {
	return srv.repository.Insert(events)
}

// Get
func (srv *Service) Get(matchId string) ([]domain.Event, error) {
	return srv.repository.Get(matchId)
}

// OnEvents
func (srv *Service) OnEvents(client domain.Client) error {
	srv.mutex.Lock()
	srv.clients[client.Id] = client
	srv.mutex.Unlock()

	return (func() error {
		<-client.Done

		srv.mutex.Lock()
		delete(srv.clients, client.Id)
		srv.mutex.Unlock()

		return nil
	})()
}
