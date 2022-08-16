package service

import (
	"context"
	"net"

	pg_dao "github.com/olegfomenko/pg-dao"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/config"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/data"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SaverService struct {
	UnimplementedSaverServer
	log          *logan.Entry
	transactions pg_dao.DAO
	listener     net.Listener
}

func NewSaverService(cfg config.Config) *SaverService {
	return &SaverService{
		log:          cfg.Log(),
		transactions: pg_dao.NewDAO(cfg.DB(), data.TransactionsTableName),
		listener:     cfg.Listener(),
	}
}

func (s *SaverService) Run() error {
	grpcServer := grpc.NewServer()
	RegisterSaverServer(grpcServer, s)
	return grpcServer.Serve(s.listener)
}

func (s *SaverService) Transactions() pg_dao.DAO {
	return s.transactions.Clone()
}

// gRPC service implementation

var _ SaverServer = &SaverService{}

func (s *SaverService) GetTransactionInfo(ctx context.Context, request *MsgTransactionInfoRequest) (*MsgTransactionInfoResponse, error) {
	tx := data.Transactions{}
	ok, err := s.Transactions().FilterByColumn(data.TransactionsHashColumn, request.Hash).Get(&tx)

	if err != nil {
		s.log.WithError(err).Error("error getting db enrty")
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	if !ok {
		return nil, status.Errorf(codes.NotFound, "Transaction not found")
	}

	return &MsgTransactionInfoResponse{
		TokenAddress:  tx.TokenAddress,
		TokenId:       tx.TokenId,
		TargetNetwork: tx.TargetNetwork,
		Receiver:      tx.Receiver,
	}, nil
}
