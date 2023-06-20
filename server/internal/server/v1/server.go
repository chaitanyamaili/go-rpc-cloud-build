package v1

import (
	"context"
	"log"
	"strings"

	v1 "github.com/chaitanyamaili/go-rpc-cloud-build/internal/types/v1"
	"github.com/chaitanyamaili/go-rpc-cloud-build/server/internal/build"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

// Service is the gRPC server implementation of the ProvisionService service.
// It also implements the server.ServiceDescription interface.
type Service struct {
	logger *zerolog.Logger
	build  *build.Handler

	v1.UnimplementedServiceServer
}

// NewService instantiates and returns a new service.
func NewService(logger *zerolog.Logger, buildHandler *build.Handler) *Service {
	serverLog := logger.With().Str("component", "grpc-server").Logger()
	return &Service{
		logger: &serverLog,
		build:  buildHandler,
	}
}

// ServiceDescription returns a reference to ProvisionService_ServiceDesc.
func (s *Service) ServiceDescription() *grpc.ServiceDesc {
	return &v1.Service_ServiceDesc
}

// CreateNewGCS returns a reference to CreateNewProjectResponse.
func (s Service) CreateNewGCS(ctx context.Context, r *v1.CreateNewGCSRequest) (*v1.CreateNewGCSResponse, error) {
	log.Printf("Received: %v", r)
	// Generate a ProjectID from the very build pkg
	gcsName := build.GenerateGCSName(ctx)
	gcsName = strings.ReplaceAll(gcsName, "-", "")
	// Trigger a Build pipeline with the specified params. Return ASAP
	buildMetadata, err := s.build.CreateNewBuild(ctx, gcsName, r.Region)
	if err != nil {
		return nil, err
	}
	return &v1.CreateNewGCSResponse{
		BuildId: buildMetadata.Build.Id,
		Status:  "success",
		Name:    gcsName,
	}, nil
}
