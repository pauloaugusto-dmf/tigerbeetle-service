package validation

import (
	"errors"

	tb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

func IsZeroID(id tb_types.Uint128) bool {
	var zero tb_types.Uint128
	return id == zero
}

// ValidateAccount checks whether required fields of an account are valid
func ValidateAccount(account tb_types.Account) error {
	if IsZeroID(account.ID) {
		return errors.New("account ID cannot be zero")
	}

	if IsZeroID(account.UserData128) {
		return errors.New("user ID cannot be zero")
	}

	if account.Ledger == 0 {
		return errors.New("ledger cannot be zero")
	}

	if account.Code == 0 {
		return errors.New("code cannot be zero")
	}

	return nil
}

func ValidateTransfer(transfer tb_types.Transfer) error {
	if IsZeroID(transfer.ID) {
		return errors.New("transfer ID cannot be zero")
	}
	if IsZeroID(transfer.DebitAccountID) || IsZeroID(transfer.CreditAccountID) {
		return errors.New("debit and credit account IDs must be set")
	}
	if IsZeroID(transfer.Amount) {
		return errors.New("transfer amount must be greater than zero")
	}
	if transfer.Ledger == 0 {
		return errors.New("ledger cannot be zero")
	}
	return nil
}
