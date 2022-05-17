package server

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bervimo/events/internal/adapters"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pb "go.buf.build/bervimo/go-grpc-gateway/bervimo/events/v1"
)

func eTag(value []byte) string {
	hash := fmt.Sprintf("%x", sha1.Sum(value))

	return fmt.Sprintf("\"%d-%s\"", len(value), hash)
}

func forwardResponse(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	bytes, _ := json.Marshal(resp)

	w.Header().Set("Cache-Control", "max-age=3600")
	w.Header().Set("ETag", eTag(bytes))

	return nil
}

func errorHandler(ctx context.Context, sm *runtime.ServeMux, ma runtime.Marshaler, rw http.ResponseWriter, req *http.Request, err error) {
	status := status.Convert(err)
	code := runtime.HTTPStatusFromCode(status.Code())

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(code)

	json.NewEncoder(rw).Encode(map[string]any{
		"code":    code,
		"message": status.Message(),
	})
}

func grpcHandlerFunc(grpcServer *grpc.Server, httHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)

			return
		}

		httHandler.ServeHTTP(w, r)
	}), &http2.Server{})
}

// NewServer
func NewServer(adapter *adapters.GRPCAdapter) (*grpc.Server, *health.Server) {
	srv := grpc.NewServer([]grpc.ServerOption{
		grpc.ConnectionTimeout(time.Duration(10) * time.Second),
		grpc.StreamInterceptor(adapters.ClientInterceptor),
	}...)

	// Reflection
	reflection.Register(srv)

	// Healthcheck
	hs := health.NewServer()

	grpc_health_v1.RegisterHealthServer(srv, hs)

	// Register rpc's
	pb.RegisterEventsServiceServer(srv, adapter)

	return srv, hs
}

// StartServer
func StartServer(srv *grpc.Server, port int) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	mux := runtime.NewServeMux(
		runtime.WithForwardResponseOption(forwardResponse),
		runtime.WithErrorHandler(errorHandler),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterEventsServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf(":%d", port), opts)

	if err != nil {
		return err
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", port), grpcHandlerFunc(srv, mux))
}

// GracefulShutdown
func GracefulShutdown(hc *health.Server, cb func(os.Signal)) {
	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	sig := <-done

	hc.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	<-time.After(time.Duration(5) * time.Second)

	// Callback handler
	cb(sig)
}
