package service

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/big"
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

func Uint128ToString(u types.Uint128) string {
	// TigerBeetle representa Uint128 como little-endian.
	// Reverte para big-endian e converte para big.Int
	reversed := make([]byte, 16)
	copy(reversed, u[:])
	// Como está little-endian, precisamos inverter
	for i := 0; i < 8; i++ {
		reversed[i], reversed[15-i] = u[15-i], u[i]
	}
	return new(big.Int).SetBytes(reversed).String()
}

func ParseUint128FromString(s string) (types.Uint128, error) {
	i := new(big.Int)
	_, ok := i.SetString(s, 10)
	if !ok {
		return types.Uint128{}, errors.New("não foi possível converter string para big.Int")
	}

	if i.Sign() < 0 {
		return types.Uint128{}, errors.New("valor negativo não é suportado para uint128")
	}

	// Obtem os bytes da representação big-endian
	b := i.Bytes()

	// Se o número for menor que 16 bytes, precisamos preenchê-lo com zeros à esquerda
	if len(b) > 16 {
		return types.Uint128{}, errors.New("valor excede 128 bits")
	}

	var u types.Uint128
	copy(u[16-len(b):], b) // insere os bytes no final, mantendo big-endian

	// Reverter para little-endian (TigerBeetle usa little-endian)
	for i := 0; i < 8; i++ {
		u[i], u[15-i] = u[15-i], u[i]
	}

	return u, nil
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
	id := tb_types.ID()

	// Usando uint64 diretamente para a ID
	account := tb_types.Account{
		ID:          id,
		UserData128: types.ToUint128(0),
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
