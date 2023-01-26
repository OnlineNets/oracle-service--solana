package grpc

import (
	"context"
	"net"

	"gitlab.com/distributed_lab/logan/v3"
	rarimotypes "gitlab.com/rarimo/rarimo-core/x/rarimocore/types"
	lib "gitlab.com/rarimo/savers/saver-grpc-lib/grpc"
	"gitlab.com/rarimo/savers/saver-grpc-lib/voter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SaverService struct {
	lib.UnimplementedSaverServer
	log      *logan.Entry
	listener net.Listener
	voter    *voter.Voter
	rarimo   *grpc.ClientConn
}

func NewSaverService(log *logan.Entry, listener net.Listener, voter *voter.Voter) *SaverService {
	return &SaverService{
		log:      log,
		listener: listener,
		voter:    voter,
	}
}

func (s *SaverService) Run() error {
	grpcServer := grpc.NewServer()
	lib.RegisterSaverServer(grpcServer, s)
	return grpcServer.Serve(s.listener)
}

// gRPC service implementation
var _ lib.SaverServer = &SaverService{}

func (s *SaverService) Revote(ctx context.Context, req *lib.RevoteRequest) (*lib.RevoteResponse, error) {
	op, err := rarimotypes.NewQueryClient(s.rarimo).Operation(ctx, &rarimotypes.QueryGetOperationRequest{Index: req.Operation})
	if err != nil {
		s.log.WithError(err).Error("error fetching op")
		return nil, status.Error(codes.Internal, "Internal error")
	}

	if err := s.voter.Process(ctx, op.Operation); err != nil {
		s.log.WithError(err).Error("error processing op")
		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &lib.RevoteResponse{}, nil
}
