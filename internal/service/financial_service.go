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

// FinancialService implementa a interface gRPC
type FinancialService struct {
	pb.UnimplementedFinancialServiceServer
	repo *repository.TigerBeetleRepository
}

// NewFinancialService cria uma nova instância do serviço
func NewFinancialService(repo *repository.TigerBeetleRepository) *FinancialService {
	return &FinancialService{
		repo: repo,
	}
}

// CreateAccount cria uma nova conta
func (s *FinancialService) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.AccountResponse, error) {
	id := tb_types.ID()

	// Usando uint64 diretamente para a ID
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
		return nil, fmt.Errorf("erro ao converter créditos: %w", err)
	}

	debits, err := Uint128ToUint64Safe(account.DebitsPosted)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter débitos: %w", err)
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

func (s *FinancialService) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.AccountResponse, error) {

	id, err := ParseUint128FromString(req.Id)
	if err != nil {
		log.Fatalf("Erro ao converter ID: %v", err)
	}

	account, err := s.repo.GetAccount(ctx, id)
	if err != nil {
		log.Printf("Erro ao buscar conta: %v", err)
		return &pb.AccountResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	account_id := Uint128ToString(account.ID)

	credits, err := Uint128ToUint64Safe(account.CreditsPosted)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter créditos: %w", err)
	}

	debits, err := Uint128ToUint64Safe(account.DebitsPosted)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter débitos: %w", err)
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

func (s *FinancialService) CreateTransfer(ctx context.Context, req *pb.CreateTransferRequest) (*pb.TransferResponse, error) {
	log.Printf("Recebida solicitação para criar transferência: %v", req)

	code, err := strconv.ParseUint(req.Code, 10, 16)
	if err != nil {
		log.Printf("Código inválido: %v", err)
		return &pb.TransferResponse{
			Success:      false,
			ErrorMessage: "Código inválido: " + err.Error(),
		}, status.Error(codes.InvalidArgument, "Código inválido")
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
		log.Printf("Erro ao criar transferência: %v", err)
		return &pb.TransferResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	idResult, err := Uint128ToUint64Safe(created.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter DebitAccountID: %w", err)
	}

	debitAccountId, err := Uint128ToUint64Safe(created.DebitAccountID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter DebitAccountID: %w", err)
	}

	creditAccountId, err := Uint128ToUint64Safe(created.CreditAccountID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter CreditAccountID: %w", err)
	}

	amounResult, err := Uint128ToUint64Safe(created.Amount)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter Amount: %w", err)
	}

	response := &pb.TransferResponse{
		Id:              idResult, // Considerando que o ID ainda precisa de um tratamento semelhante se for um Uint128
		DebitAccountId:  debitAccountId,
		CreditAccountId: creditAccountId,
		Ledger:          created.Ledger,
		Amount:          amounResult, // Convertendo o valor de uint64 para int64, caso necessário
		Code:            strconv.Itoa(int(created.Code)),
		Flags:           uint32(created.Flags),
		Success:         true,
	}

	return response, nil
}

func (s *FinancialService) GetTransfer(ctx context.Context, req *pb.GetTransferRequest) (*pb.TransferResponse, error) {
	log.Printf("Recebida solicitação para buscar transferência: %v", req)

	id := tb_types.ToUint128(req.Id)

	transfer, err := s.repo.GetTransfer(ctx, id)
	if err != nil {
		log.Printf("Erro ao buscar transferência: %v", err)
		return &pb.TransferResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	idResult, err := Uint128ToUint64Safe(transfer.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter DebitAccountID: %w", err)
	}

	debitAccountId, err := Uint128ToUint64Safe(transfer.DebitAccountID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter DebitAccountID: %w", err)
	}

	creditAccountId, err := Uint128ToUint64Safe(transfer.CreditAccountID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter CreditAccountID: %w", err)
	}

	amounResult, err := Uint128ToUint64Safe(transfer.Amount)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter Amount: %w", err)
	}

	response := &pb.TransferResponse{
		Id:              idResult,
		DebitAccountId:  debitAccountId,
		CreditAccountId: creditAccountId,
		Ledger:          transfer.Ledger,
		Amount:          amounResult,
		Code:            strconv.Itoa(int(transfer.Code)),
		Flags:           uint32(transfer.Flags),
		Success:         true,
	}

	return response, nil
}
