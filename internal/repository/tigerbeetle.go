package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/pauloaugusto-dmf/tigerbeetle-service/internal/logger"
	"github.com/pauloaugusto-dmf/tigerbeetle-service/internal/validation"

	tb "github.com/tigerbeetle/tigerbeetle-go"
	tb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type TigerBeetleRepository struct {
	client tb.Client
}

func NewTigerBeetleRepository(addresses []string, clusterID uint64) (*TigerBeetleRepository, error) {
	client, err := tb.NewClient(tb_types.ToUint128(clusterID), addresses)
	if err != nil {
		return nil, fmt.Errorf("failed to create TigerBeetle client: %w", err)
	}

	return &TigerBeetleRepository{
		client: client,
	}, nil
}

func (r *TigerBeetleRepository) Close() {
	if r.client != nil {
		r.client.Close()
	}
}

func (r *TigerBeetleRepository) CreateAccount(ctx context.Context, account tb_types.Account) ([]tb_types.AccountEventResult, error) {
	if err := validation.ValidateAccount(account); err != nil {
		logger.Error("account validation failed", "error", err)
		return nil, err
	}

	logger.Info("creating account", "id", account.ID, "ledger", account.Ledger)

	results, err := r.client.CreateAccounts([]tb_types.Account{account})
	if err != nil {
		logger.Error("error creating account", "error", err)
		return nil, err
	}

	for _, result := range results {
		if result.Result != 0 {
			logger.Error("account creation failed", "result_code", result.Result, "id", account.ID)
			return results, fmt.Errorf("account creation failed with code %d", result.Result)
		}
	}

	return results, nil
}

func (r *TigerBeetleRepository) GetAccount(ctx context.Context, id tb_types.Uint128) (*tb_types.Account, error) {
	logger.Debug("looking up account", "id", id)

	accounts, err := r.client.LookupAccounts([]tb_types.Uint128{id})
	if err != nil {
		logger.Error("failed to fetch account", "error", err)
		return nil, fmt.Errorf("failed to fetch account: %w", err)
	}
	if len(accounts) == 0 {
		logger.Info("account not found", "id", id)
		return nil, errors.New("account not found")
	}

	return &accounts[0], nil
}

func (r *TigerBeetleRepository) CreateTransfer(ctx context.Context, transfer tb_types.Transfer) (*tb_types.Transfer, error) {
	if err := validation.ValidateTransfer(transfer); err != nil {
		logger.Error("transfer validation failed", "error", err)
		return nil, err
	}

	logger.Info("creating transfer", "id", transfer.ID, "amount", transfer.Amount)

	_, err := r.client.CreateTransfers([]tb_types.Transfer{transfer})
	if err != nil {
		logger.Error("error creating transfer", "error", err)
		return nil, fmt.Errorf("failed to create transfer: %w", err)
	}

	return &transfer, nil
}

func (r *TigerBeetleRepository) GetTransfer(ctx context.Context, id tb_types.Uint128) (*tb_types.Transfer, error) {
	logger.Debug("looking up transfer", "id", id)

	transfers, err := r.client.LookupTransfers([]tb_types.Uint128{id})
	if err != nil {
		logger.Error("failed to fetch transfer", "error", err)
		return nil, fmt.Errorf("failed to fetch transfer: %w", err)
	}
	if len(transfers) == 0 {
		logger.Info("transfer not found", "id", id)
		return nil, errors.New("transfer not found")
	}

	return &transfers[0], nil
}
