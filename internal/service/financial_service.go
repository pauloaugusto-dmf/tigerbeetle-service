package service

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/pauloaugusto-dmf/tigerbeetle-service/internal/repository"

	. "github.com/pauloaugusto-dmf/tigerbeetle-service/internal/tbutil"
	pb "github.com/pauloaugusto-dmf/tigerbeetle-service/proto"
	tb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FinancialService implements the gRPC interface
type FinancialService struct {
	pb.UnimplementedFinancialServiceServer
	repo *repository.TigerBeetleRepository
}

// NewFinancialService creates a new instance of the service
func NewFinancialService(repo *repository.TigerBeetleRepository) *FinancialService {
	return &FinancialService{
		repo: repo,
	}
}

// CreateAccount creates a new account
func (s *FinancialService) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.AccountResponse, error) {
	id := tb_types.ID()

	account := tb_types.Account{
		ID:          id,
		UserData128: tb_types.ToUint128(0),
		UserData64:  0,
		UserData32:  0,
		Ledger:      req.Ledger,
		Code:        uint16(req.Code),
		Flags:       uint16(req.Flags),
		Timestamp:   0,
	}

	_, err := s.repo.CreateAccount(ctx, account)
	if err != nil {
		return &pb.AccountResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	credits, err := Uint128ToUint64Safe(account.CreditsPosted)
	if err != nil {
		return nil, fmt.Errorf("error converting credits: %w", err)
	}

	debits, err := Uint128ToUint64Safe(account.DebitsPosted)
	if err != nil {
		return nil, fmt.Errorf("error converting debits: %w", err)
	}

	balance := credits - debits
	account_id := Uint128ToString(account.ID)

	response := &pb.AccountResponse{
		Id:       account_id,
		Code:     uint32(account.Code),
		Ledger:   account.Ledger,
		Balance:  int64(balance),
		Flags:    uint32(account.Flags),
		UserData: "",
		Success:  true,
	}

	return response, nil
}

// GetAccount fetches an account by ID
func (s *FinancialService) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.AccountResponse, error) {
	id, err := ParseUint128FromString(req.Id)
	if err != nil {
		log.Fatalf("Error converting ID: %v", err)
	}

	account, err := s.repo.GetAccount(ctx, id)
	if err != nil {
		log.Printf("Error fetching account: %v", err)
		return &pb.AccountResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	account_id := Uint128ToString(account.ID)

	credits, err := Uint128ToUint64Safe(account.CreditsPosted)
	if err != nil {
		return nil, fmt.Errorf("error converting credits: %w", err)
	}

	debits, err := Uint128ToUint64Safe(account.DebitsPosted)
	if err != nil {
		return nil, fmt.Errorf("error converting debits: %w", err)
	}

	balance := credits - debits

	response := &pb.AccountResponse{
		Id:      account_id,
		Code:    uint32(account.Code),
		Ledger:  account.Ledger,
		Balance: int64(balance),
		Flags:   uint32(account.Flags),
		Success: true,
	}

	return response, nil
}

// CreateTransfer creates a new transfer
func (s *FinancialService) CreateTransfer(ctx context.Context, req *pb.CreateTransferRequest) (*pb.TransferResponse, error) {
	log.Printf("Received request to create transfer: %v", req)

	code, err := strconv.ParseUint(req.Code, 10, 16)
	if err != nil {
		log.Printf("Invalid code: %v", err)
		return &pb.TransferResponse{
			Success:      false,
			ErrorMessage: "Invalid code: " + err.Error(),
		}, status.Error(codes.InvalidArgument, "Invalid code")
	}

	debit_account_id := tb_types.ToUint128(req.DebitAccountId)
	credit_account_id := tb_types.ToUint128(req.CreditAccountId)
	amount := tb_types.ToUint128(req.Amount)

	transfer := tb_types.Transfer{
		ID:              tb_types.ID(),
		DebitAccountID:  debit_account_id,
		CreditAccountID: credit_account_id,
		Ledger:          req.Ledger,
		Code:            uint16(code),
		Flags:           uint16(req.Flags),
		Amount:          amount,
	}

	created, err := s.repo.CreateTransfer(ctx, transfer)
	if err != nil {
		log.Printf("Error creating transfer: %v", err)
		return &pb.TransferResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	idResult, err := Uint128ToUint64Safe(created.ID)
	if err != nil {
		return nil, fmt.Errorf("error converting ID: %w", err)
	}

	debitAccountId, err := Uint128ToUint64Safe(created.DebitAccountID)
	if err != nil {
		return nil, fmt.Errorf("error converting DebitAccountID: %w", err)
	}

	creditAccountId, err := Uint128ToUint64Safe(created.CreditAccountID)
	if err != nil {
		return nil, fmt.Errorf("error converting CreditAccountID: %w", err)
	}

	amountResult, err := Uint128ToUint64Safe(created.Amount)
	if err != nil {
		return nil, fmt.Errorf("error converting Amount: %w", err)
	}

	response := &pb.TransferResponse{
		Id:              idResult,
		DebitAccountId:  debitAccountId,
		CreditAccountId: creditAccountId,
		Ledger:          created.Ledger,
		Amount:          amountResult,
		Code:            strconv.Itoa(int(created.Code)),
		Flags:           uint32(created.Flags),
		Success:         true,
	}

	return response, nil
}

// GetTransfer fetches a transfer by ID
func (s *FinancialService) GetTransfer(ctx context.Context, req *pb.GetTransferRequest) (*pb.TransferResponse, error) {
	log.Printf("Received request to fetch transfer: %v", req)

	id := tb_types.ToUint128(req.Id)

	transfer, err := s.repo.GetTransfer(ctx, id)
	if err != nil {
		log.Printf("Error fetching transfer: %v", err)
		return &pb.TransferResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	idResult, err := Uint128ToUint64Safe(transfer.ID)
	if err != nil {
		return nil, fmt.Errorf("error converting ID: %w", err)
	}

	debitAccountId, err := Uint128ToUint64Safe(transfer.DebitAccountID)
	if err != nil {
		return nil, fmt.Errorf("error converting DebitAccountID: %w", err)
	}

	creditAccountId, err := Uint128ToUint64Safe(transfer.CreditAccountID)
	if err != nil {
		return nil, fmt.Errorf("error converting CreditAccountID: %w", err)
	}

	amountResult, err := Uint128ToUint64Safe(transfer.Amount)
	if err != nil {
		return nil, fmt.Errorf("error converting Amount: %w", err)
	}

	response := &pb.TransferResponse{
		Id:              idResult,
		DebitAccountId:  debitAccountId,
		CreditAccountId: creditAccountId,
		Ledger:          transfer.Ledger,
		Amount:          amountResult,
		Code:            strconv.Itoa(int(transfer.Code)),
		Flags:           uint32(transfer.Flags),
		Success:         true,
	}

	return response, nil
}
