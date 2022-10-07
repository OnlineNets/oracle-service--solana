package service

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gagliardetto/solana-go"
	"gitlab.com/distributed_lab/logan/v3"
	lib "gitlab.com/rarify-protocol/saver-grpc-lib/grpc"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/config"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/data/pg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SaverService struct {
	lib.UnimplementedSaverServer
	log      *logan.Entry
	storage  *pg.Storage
	listener net.Listener
}

func NewSaverService(cfg config.Config) *SaverService {
	return &SaverService{
		log:      cfg.Log(),
		storage:  cfg.Storage(),
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

func (s *SaverService) GetDepositInfo(ctx context.Context, request *lib.MsgTransactionInfoRequest) (*lib.MsgDepositResponse, error) {
	s.log.Infof("[GET DEPOSIT] request: hash=%s event_id=%s token_type=%d", request.Hash, request.EventId, request.Type)

	instructionId, err := strconv.Atoi(request.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Wrong event id format")
	}

	switch TokenType(request.Type) {
	case TypeNative:
		return s.getNativeDeposit(ctx, request.Hash, instructionId)
	case TypeFT:
		return s.getFTDeposit(ctx, request.Hash, instructionId)
	case TypeNFT:
		return s.getNFTDeposit(ctx, request.Hash, instructionId)
	}
	return nil, status.Errorf(codes.InvalidArgument, "Wrong token type")
}

func (s *SaverService) getNativeDeposit(ctx context.Context, hash string, instructionId int) (*lib.MsgDepositResponse, error) {
	entry, err := s.storage.Clone().NativeDepositQ().NativeDepositByHashInstructionIDCtx(ctx, hash, instructionId, false)
	if err != nil {
		s.log.WithError(err).Error("error getting database entry")
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	if entry == nil {
		return nil, status.Errorf(codes.NotFound, "Deposit not found")
	}

	return &lib.MsgDepositResponse{
		TargetNetwork: entry.TargetNetwork,
		Sender:        hexutil.Encode(solana.MustPublicKeyFromBase58(entry.Sender).Bytes()),
		Receiver:      entry.Receiver,
		Amount:        fmt.Sprint(entry.Amount),
		BundleData:    entry.BundleData.String,
		BundleSalt:    entry.BundleSeed.String,
	}, nil
}

func (s *SaverService) getFTDeposit(ctx context.Context, hash string, instructionId int) (*lib.MsgDepositResponse, error) {
	entry, err := s.storage.Clone().FtDepositQ().FtDepositByHashInstructionIDCtx(ctx, hash, instructionId, false)
	if err != nil {
		s.log.WithError(err).Error("error getting database entry")
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	if entry == nil {
		return nil, status.Errorf(codes.NotFound, "Deposit not found")
	}

	return &lib.MsgDepositResponse{
		TargetNetwork: entry.TargetNetwork,
		Sender:        hexutil.Encode(solana.MustPublicKeyFromBase58(entry.Sender).Bytes()),
		Receiver:      entry.Receiver,
		Amount:        fmt.Sprint(entry.Amount),
		TokenAddress:  hexutil.Encode(solana.MustPublicKeyFromBase58(entry.Mint).Bytes()),
		BundleData:    entry.BundleData.String,
		BundleSalt:    entry.BundleSeed.String,
	}, nil
}

func (s *SaverService) getNFTDeposit(ctx context.Context, hash string, instructionId int) (*lib.MsgDepositResponse, error) {
	entry, err := s.storage.Clone().NftDepositQ().NftDepositByHashInstructionIDCtx(ctx, hash, instructionId, false)
	if err != nil {
		s.log.WithError(err).Error("error getting database entry")
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	if entry == nil {
		return nil, status.Errorf(codes.NotFound, "Deposit not found")
	}

	collection := ""
	if entry.Collection.Valid {
		collection = hexutil.Encode(solana.MustPublicKeyFromBase58(entry.Collection.String).Bytes())
	}

	return &lib.MsgDepositResponse{
		TargetNetwork: entry.TargetNetwork,
		Sender:        hexutil.Encode(solana.MustPublicKeyFromBase58(entry.Sender).Bytes()),
		Receiver:      entry.Receiver,
		TokenAddress:  collection,
		TokenId:       hexutil.Encode(solana.MustPublicKeyFromBase58(entry.Mint).Bytes()),
		BundleData:    entry.BundleData.String,
		BundleSalt:    entry.BundleSeed.String,
	}, nil
}
