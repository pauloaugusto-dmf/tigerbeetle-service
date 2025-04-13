package service

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"strconv"

	pb "github.com/pauloaugusto-dmf/tigerbeetle-service/proto"

	"github.com/pauloaugusto-dmf/tigerbeetle-service/internal/repository"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	tb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Uint128ToUint64Safe(u types.Uint128) (uint64, error) {
	// Verifica os bytes mais significativos (8 a 15)
	for i := 8; i < 16; i++ {
		if u[i] != 0 {
			return 0, errors.New("valor excede uint64")
		}
	}

	// Converte os primeiros 8 bytes (parte baixa)
	return binary.LittleEndian.Uint64(u[0:8]), nil
}

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
	log.Printf("Recebida solicitação para criar conta: %v", req)

	// Usando uint64 diretamente para a ID
	account := tb_types.Account{
		ID:     tb_types.ID(),
		Ledger: req.Ledger,
		Code:   uint16(req.Code),
		Flags:  uint16(req.Flags),
	}

	result, err := s.repo.CreateAccount(ctx, account)
	if err != nil {
		log.Printf("Erro ao criar conta: %v", err)
		return &pb.AccountResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	if result != tb_types.AccountOK {
		msg := "Erro ao criar conta: " + result.String()
		log.Println(msg)
		return &pb.AccountResponse{
			Success:      false,
			ErrorMessage: msg,
		}, status.Error(codes.Internal, msg)
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

	account_id, err := Uint128ToUint64Safe(account.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID: %w", err)
	}

	response := &pb.AccountResponse{
		Id:       account_id,           // Agora usamos diretamente o ID como uint64
		Code:     uint32(account.Code), // Converta para string se necessário
		Ledger:   account.Ledger,
		Balance:  int64(balance),
		Flags:    uint32(account.Flags),
		UserData: "", // Se precisar, adapte aqui
		Success:  true,
	}

	return response, nil
}

func (s *FinancialService) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.AccountResponse, error) {
	log.Printf("Recebida solicitação para buscar conta: %v", req)

	id := tb_types.ToUint128(req.Id)

	account, err := s.repo.GetAccount(ctx, id)
	if err != nil {
		log.Printf("Erro ao buscar conta: %v", err)
		return &pb.AccountResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	account_id, err := Uint128ToUint64Safe(account.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID: %w", err)
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
