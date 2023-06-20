package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp/reuseport"
	"google.golang.org/grpc"
)

// ServiceDescriptor interface
type ServiceDescriptor interface {
	ServiceDescription() *grpc.ServiceDesc
}

// HTTPServer params for the HTTP server.
type HTTPServer struct {
	config *Config
	//router     *routing.Router
	backServer *grpc.Server
	log        *zerolog.Logger
	options    []grpc.ServerOption
	services   []ServiceDescriptor
}

// Config HTTP Configuration
type Config struct {
	Addr string
	Port int
}

// NewServer instantiates the server.
func NewServer(config *Config, logger *zerolog.Logger, opts []grpc.ServerOption) *HTTPServer {
	serverLog := logger.With().Str("component", "http-server").Logger()
	return &HTTPServer{
		config: config,
		log:    &serverLog,
	}
}

// ListenAndServe starts the server and listens
func (s *HTTPServer) ListenAndServe() error {
	addr := s.config.Addr + ":" + strconv.Itoa(s.config.Port)
	ln, err := reuseport.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	s.backServer = grpc.NewServer(s.options...)

	for _, service := range s.services {
		desc := service.ServiceDescription()
		s.backServer.RegisterService(desc, service)
		s.log.Debug().Str("name", desc.ServiceName).Msg("gRPC service registered")
	}

	s.log.Debug().Int("port", s.config.Port).Str("address", s.config.Addr).Msg("Server listening and ready")
	err = s.backServer.Serve(ln)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.log.Error().Err(err).Stack().Msg("Error occurred while trying to close server gracefully")
		return err
	}

	s.log.Info().Msg("Server closed successfully")
	return nil
}

// RegisterGRPCService appends the given service.
func (s *HTTPServer) RegisterGRPCService(service ServiceDescriptor) {
	s.services = append(s.services, service)
}

// Finish stop the server.
func (s *HTTPServer) Finish(ctx context.Context) error {
	s.log.Debug().Msg("Trying to close server")
	s.backServer.GracefulStop()
	return nil
}
