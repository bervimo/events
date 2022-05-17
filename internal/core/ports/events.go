package ports

import "github.com/bervimo/events/internal/core/domain"

type Service interface {
	Insert(events []domain.Event) (bool, error)
	Get(matchId string) ([]domain.Event, error)
	OnEvents(client domain.Client) error
}

type Repository interface {
	Insert(events []domain.Event) (bool, error)
	Get(matchId string) ([]domain.Event, error)
	OnEvents(callback func(*domain.Event, error))

	Close() error
}
