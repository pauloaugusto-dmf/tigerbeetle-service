// internal/repository/tigerbeetle.go
package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	tb "github.com/tigerbeetle/tigerbeetle-go"
	tb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// TigerBeetleRepository gerencia a conexão com o TigerBeetle
type TigerBeetleRepository struct {
	client tb.Client
}

// NewTigerBeetleRepository cria uma nova instância do repositório
func NewTigerBeetleRepository(addresses []string, clusterID uint64) (*TigerBeetleRepository, error) {
	// Cria um cliente TigerBeetle
	client, err := tb.NewClient(tb_types.ToUint128(clusterID), addresses)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar cliente TigerBeetle: %w", err)
	}

	return &TigerBeetleRepository{
		client: client,
	}, nil
}

// Close fecha a conexão com o TigerBeetle
func (r *TigerBeetleRepository) Close() {
	if r.client != nil {
		r.client.Close()
	}
}

// CreateAccount cria uma nova conta no TigerBeetle
func (r *TigerBeetleRepository) CreateAccount(ctx context.Context, account tb_types.Account) ([]tb_types.AccountEventResult, error) {
	accounts := []tb_types.Account{account}

	account = accounts[0]

	// Chamar com apenas um argumento
	results, err := r.client.CreateAccounts(accounts)
	log.Printf("Result: %v", results)
	log.Printf("Result: %v", err)

	return results, nil
}

// GetAccount busca uma conta pelo ID
func (r *TigerBeetleRepository) GetAccount(ctx context.Context, id tb_types.Uint128) (*tb_types.Account, error) {
	ids := []tb_types.Uint128{id}

	// Agora recebe diretamente as contas e o erro
	accounts, err := r.client.LookupAccounts(ids)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar conta: %w", err)
	}

	if len(accounts) == 0 {
		return nil, errors.New("conta não encontrada")
	}

	return &accounts[0], nil
}

// CreateTransfer cria uma nova transferência
func (r *TigerBeetleRepository) CreateTransfer(ctx context.Context, transfer tb_types.Transfer) (*tb_types.Transfer, error) {
	transfers := []tb_types.Transfer{transfer}

	results, err := r.client.CreateTransfers(transfers)
	log.Printf("Erro: %v", err)
	log.Printf("Erro: %v", results)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar transferência: %w", err)
	}

	if results[0].Result != tb_types.TransferOK {
		return nil, fmt.Errorf("erro ao criar transferência: %s", results[0].Result.String())
	}

	return &transfer, nil
}

func (r *TigerBeetleRepository) GetTransfer(ctx context.Context, id tb_types.Uint128) (*tb_types.Transfer, error) {
	ids := []tb_types.Uint128{id}

	// Agora recebe os transfers diretamente
	transfers, err := r.client.LookupTransfers(ids)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar transferência: %w", err)
	}

	if len(transfers) == 0 {
		return nil, errors.New("transferência não encontrada")
	}

	return &transfers[0], nil
}
