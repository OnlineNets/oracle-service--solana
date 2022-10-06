package service

import (
	"context"
	"fmt"
	"net"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gagliardetto/solana-go"
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

func (s *SaverService) GetDepositInfo(ctx context.Context, request *lib.MsgTransactionInfoRequest) (*lib.MsgDepositResponse, error) {
	s.log.Infof("[GET DEPOSIT] request: hash=%s event_id=%s token_type=%d", request.Hash, request.EventId, request.Type)
	switch TokenType(request.Type) {
	case TypeNative:
		return s.getNativeDeposit(request)
	case TypeFT:
		return s.getFTDeposit(request)
	case TypeNFT:
		return s.getNFTDeposit(request)
	}
	return nil, status.Errorf(codes.InvalidArgument, "Wrong token type")
}

func (s *SaverService) getNativeDeposit(request *lib.MsgTransactionInfoRequest) (*lib.MsgDepositResponse, error) {
	entry := data.NativeDeposit{}
	ok, err := s.nativeDeposits.Clone().
		FilterByColumn(data.HashColumnName, request.Hash).
		FilterByColumn(data.InstructionIdColumnName, request.EventId).
		Get(&entry)
	if err != nil {
		s.log.WithError(err).Error("error getting database entry")
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	if !ok {
		return nil, status.Errorf(codes.NotFound, "Deposit not found")
	}

	return &lib.MsgDepositResponse{
		TargetNetwork: entry.TargetNetwork,
		Sender:        hexutil.Encode(solana.MustPublicKeyFromBase58(entry.Sender).Bytes()),
		Receiver:      entry.Receiver,
		Amount:        fmt.Sprint(entry.Amount),
		BundleData:    entry.BundleData,
		BundleSalt:    entry.BundleSeed,
	}, nil
}

func (s *SaverService) getFTDeposit(request *lib.MsgTransactionInfoRequest) (*lib.MsgDepositResponse, error) {
	entry := data.FTDeposit{}
	ok, err := s.ftDeposits.Clone().
		FilterByColumn(data.HashColumnName, request.Hash).
		FilterByColumn(data.InstructionIdColumnName, request.EventId).
		Get(&entry)
	if err != nil {
		s.log.WithError(err).Error("error getting database entry")
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	if !ok {
		return nil, status.Errorf(codes.NotFound, "Deposit not found")
	}

	return &lib.MsgDepositResponse{
		TargetNetwork: entry.TargetNetwork,
		Sender:        hexutil.Encode(solana.MustPublicKeyFromBase58(entry.Sender).Bytes()),
		Receiver:      entry.Receiver,
		Amount:        fmt.Sprint(entry.Amount),
		TokenAddress:  hexutil.Encode(solana.MustPublicKeyFromBase58(entry.Mint).Bytes()),
		BundleData:    entry.BundleData,
		BundleSalt:    entry.BundleSeed,
	}, nil
}

func (s *SaverService) getNFTDeposit(request *lib.MsgTransactionInfoRequest) (*lib.MsgDepositResponse, error) {
	entry := data.NFTDeposit{}
	ok, err := s.nftDeposits.Clone().
		FilterByColumn(data.HashColumnName, request.Hash).
		FilterByColumn(data.InstructionIdColumnName, request.EventId).
		Get(&entry)

	if err != nil {
		s.log.WithError(err).Error("error getting database entry")
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	if !ok {
		return nil, status.Errorf(codes.NotFound, "Deposit not found")
	}

	return &lib.MsgDepositResponse{
		TargetNetwork: entry.TargetNetwork,
		Sender:        hexutil.Encode(solana.MustPublicKeyFromBase58(entry.Sender).Bytes()),
		Receiver:      entry.Receiver,
		TokenAddress:  hexutil.Encode(solana.MustPublicKeyFromBase58(entry.Collection).Bytes()),
		TokenId:       hexutil.Encode(solana.MustPublicKeyFromBase58(entry.Mint).Bytes()),
		BundleData:    entry.BundleData,
		BundleSalt:    entry.BundleSeed,
	}, nil
}
