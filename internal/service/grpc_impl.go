package service

import (
	"context"
	"net"

	pg_dao "github.com/olegfomenko/pg-dao"
	"gitlab.com/distributed_lab/logan/v3"
	lib "gitlab.com/rarify-protocol/saver-grpc-lib/grpc"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/config"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/data"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SaverService struct {
	lib.UnimplementedSaverServer
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
	lib.RegisterSaverServer(grpcServer, s)
	return grpcServer.Serve(s.listener)
}

func (s *SaverService) Transactions() pg_dao.DAO {
	return s.transactions.Clone()
}

// gRPC service implementation

var _ lib.SaverServer = &SaverService{}

func (s *SaverService) GetTransactionInfo(ctx context.Context, request *lib.MsgTransactionInfoRequest) (*lib.MsgTransactionInfoResponse, error) {
	tx := data.Transaction{}
	ok, err := s.Transactions().FilterByColumn(data.TransactionsHashColumn, request.Hash).Get(&tx)

	if err != nil {
		s.log.WithError(err).Error("error getting db enrty")
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	if !ok {
		return nil, status.Errorf(codes.NotFound, "Transaction not found")
	}

	return &lib.MsgTransactionInfoResponse{
		TokenId:       tx.TokenId,
		Collection:    tx.Collection,
		TargetNetwork: tx.TargetNetwork,
		Receiver:      tx.Receiver,
		TokenType:     lib.Type(tx.TokenType),
	}, nil
}
