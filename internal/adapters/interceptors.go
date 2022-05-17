package adapters

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type wrappedStream struct {
	context context.Context
	grpc.ServerStream
}

func (ws *wrappedStream) Context() context.Context {
	return ws.context
}

func newWrappedStream(ss grpc.ServerStream, ctx context.Context) grpc.ServerStream {
	return &wrappedStream{ServerStream: ss, context: ctx}
}

const (
	contextClientIdKey  = "client_id"
	metadataClientIdKey = "x-client-id"
)

// Intercept
func ClientInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	md, _ := metadata.FromIncomingContext(ss.Context())

	values := md.Get(metadataClientIdKey)

	if len(values) == 0 {
		return status.Error(codes.InvalidArgument, "Invalid argument")
	}

	clientId := values[0]

	ctx := context.WithValue(ss.Context(), contextClientIdKey, clientId)

	return handler(srv, newWrappedStream(ss, ctx))
}
