package adapters

import (
	"context"

	"github.com/bervimo/events/internal/core/domain"
	"github.com/bervimo/events/internal/core/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "go.buf.build/bervimo/go-grpc-gateway/bervimo/events/v1"
)

// GRPCAdapter
type GRPCAdapter struct {
	service ports.Service
	pb.UnimplementedEventsServiceServer
}

// NewGRPCAdapter
func NewGRPCAdapter(service ports.Service) *GRPCAdapter {
	return &GRPCAdapter{service: service}
}

// Insert
func (ad *GRPCAdapter) Insert(ctx context.Context, req *pb.InsertRequest) (*pb.InsertResponse, error) {
	err := req.Validate()

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	events := make([]domain.Event, len(req.Events))

	for i, event := range req.Events {
		events[i] = domain.Event{
			Id:      event.Id,
			MatchId: event.MatchId,
			Code:    domain.Code(event.Code),
			Head:    event.Head,
			Next:    event.Next,
		}
	}

	res, err := ad.service.Insert(events)

	if err != nil {
		return nil, err
	}

	return &pb.InsertResponse{Inserted: res}, nil
}

// Get
func (ad *GRPCAdapter) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	err := req.Validate()

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := ad.service.Get(req.MatchId)

	if err != nil {
		return nil, err
	}

	events := make([]*pb.Event, len(res))

	for i, event := range res {
		events[i] = &pb.Event{
			Id:      event.Id,
			MatchId: event.MatchId,
			Code:    pb.Code(event.Code),
			Head:    event.Head,
			Next:    event.Next,
		}
	}

	return &pb.GetResponse{Events: events}, nil
}

// OnEvents
func (ad *GRPCAdapter) OnEvents(req *pb.OnEventsRequest, str pb.EventsService_OnEventsServer) error {
	id := str.Context().Value(contextClientIdKey).(string)

	cli := domain.Client{
		Id: id,
		Callback: func(event *domain.Event, err error) {
			item := &pb.Event{
				Id:      event.Id,
				MatchId: event.MatchId,
				Code:    pb.Code(event.Code),
				Next:    event.Next,
			}

			str.Send(&pb.OnEventsResponse{Event: item})
		},
		Done: str.Context().Done(),
	}

	return ad.service.OnEvents(cli)
}
