package usersvc

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type GrpcHealthChecker struct{}

func NewGrpHealthChecker() *GrpcHealthChecker {
	return &GrpcHealthChecker{}
}

func (s *GrpcHealthChecker) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (h *GrpcHealthChecker) Watch(*grpc_health_v1.HealthCheckRequest, grpc_health_v1.Health_WatchServer) error {
	return nil
}

func NewHttpHealthCheckerHandler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/health", healthCheckHandler).Methods("GET")
	return r
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
