package service

import (
	"context"
	"math/big"
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
	log            *logan.Entry
	nativeDeposits pg_dao.DAO
	ftDeposits     pg_dao.DAO
	nftDeposits    pg_dao.DAO
	listener       net.Listener
}

func NewSaverService(cfg config.Config) *SaverService {
	return &SaverService{
		log:            cfg.Log(),
		nativeDeposits: pg_dao.NewDAO(cfg.DB(), data.NativeDepositsTableName),
		ftDeposits:     pg_dao.NewDAO(cfg.DB(), data.FTDepositsTableName),
		nftDeposits:    pg_dao.NewDAO(cfg.DB(), data.NFTDepositsTableName),
		listener:       cfg.Listener(),
	}
}

func (s *SaverService) Run() error {
	grpcServer := grpc.NewServer()
	lib.RegisterSaverServer(grpcServer, s)
	return grpcServer.Serve(s.listener)
}

// gRPC service implementation

var _ lib.SaverServer = &SaverService{}

func (s *SaverService) GetNativeDepositInfo(ctx context.Context, request *lib.MsgTransactionInfoRequest) (*lib.MsgNativeDepositResponse, error) {
	entry := data.NativeDeposit{}
	ok, err := s.nativeDeposits.Clone().FilterByColumn(data.HashColumnName, request.Hash).Get(&entry)
	if err != nil {
		s.log.WithError(err).Error("error getting database entry")
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	if !ok {
		return nil, status.Errorf(codes.NotFound, "Deposit not found")
	}

	return &lib.MsgNativeDepositResponse{
		TargetNetwork: entry.TargetNetwork,
		Sender:        entry.Sender,
		Receiver:      entry.Receiver,
		Amount:        new(big.Int).SetUint64(entry.Amount).Bytes(),
	}, nil
}

func (s *SaverService) GetFTDepositInfo(ctx context.Context, request *lib.MsgTransactionInfoRequest) (*lib.MsgFTDepositResponse, error) {
	entry := data.FTDeposit{}
	ok, err := s.ftDeposits.Clone().FilterByColumn(data.HashColumnName, request.Hash).Get(&entry)
	if err != nil {
		s.log.WithError(err).Error("error getting database entry")
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	if !ok {
		return nil, status.Errorf(codes.NotFound, "Deposit not found")
	}

	return &lib.MsgFTDepositResponse{
		TargetNetwork: entry.TargetNetwork,
		Sender:        entry.Sender,
		Receiver:      entry.Receiver,
		Amount:        new(big.Int).SetUint64(entry.Amount).Bytes(),
		TokenAddress:  entry.Mint,
	}, nil
}

func (s *SaverService) GetNFTDepositInfo(ctx context.Context, request *lib.MsgTransactionInfoRequest) (*lib.MsgNFTDepositResponse, error) {
	entry := data.NFTDeposit{}
	ok, err := s.nftDeposits.Clone().FilterByColumn(data.HashColumnName, request.Hash).Get(&entry)
	if err != nil {
		s.log.WithError(err).Error("error getting database entry")
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	if !ok {
		return nil, status.Errorf(codes.NotFound, "Deposit not found")
	}

	return &lib.MsgNFTDepositResponse{
		TargetNetwork: entry.TargetNetwork,
		Sender:        entry.Sender,
		Receiver:      entry.Receiver,
		TokenAddress:  entry.Collection,
		TokenId:       entry.Mint,
	}, nil
}
