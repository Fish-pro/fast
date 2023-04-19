package v1

import (
	"context"

	apiv1 "github.com/fast-io/fast/api/proto/v1"
)

type HealthService struct {
	apiv1.UnimplementedHealthServiceServer
}

func NewHealthService() apiv1.HealthServiceServer {
	return &HealthService{}
}

func (s *HealthService) Health(context.Context, *apiv1.HealthRequest) (*apiv1.HealthResponse, error) {
	return &apiv1.HealthResponse{Msg: "ok"}, nil
}
