package grpc

import (
	"context"
	"net"

	"gitlab.com/distributed_lab/logan/v3"
	lib "gitlab.com/rarimo/savers/saver-grpc-lib/grpc"
	"gitlab.com/rarimo/savers/sol-saver-svc/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SaverService struct {
	lib.UnimplementedSaverServer
	log      *logan.Entry
	listener net.Listener
}

func NewSaverService(cfg config.Config) *SaverService {
	return &SaverService{
		log:      cfg.Log(),
		listener: cfg.Listener(),
	}
}

func (s *SaverService) Run() error {
	grpcServer := grpc.NewServer()
	lib.RegisterSaverServer(grpcServer, s)
	return grpcServer.Serve(s.listener)
}

// gRPC service implementation
var _ lib.SaverServer = &SaverService{}

func (s *SaverService) Revote(context.Context, *lib.RevoteRequest) (*lib.RevoteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Revote not implemented")
}
