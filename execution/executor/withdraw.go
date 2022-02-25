package executor

import (
	"github.com/zarbchain/zarb-go/errors"
	"github.com/zarbchain/zarb-go/sandbox"
	"github.com/zarbchain/zarb-go/tx"
	"github.com/zarbchain/zarb-go/tx/payload"
)

type WithdrawExecutor struct {
	fee    int64
	strict bool
}

func NewWithdrawExecutor(strict bool) *WithdrawExecutor {
	return &WithdrawExecutor{strict: strict}
}

func (e *WithdrawExecutor) Execute(trx *tx.Tx, sb sandbox.Sandbox) error {
	pld := trx.Payload().(*payload.WithdrawPayload)

	withdrawingVal := sb.Validator(pld.From)
	if withdrawingVal == nil {
		return errors.Errorf(errors.ErrInvalidAddress, "Unable to retrieve validator account")
	}

	if withdrawingVal.Sequence()+1 != trx.Sequence() {
		return errors.Errorf(errors.ErrInvalidSequence, "Invalid sequence, Expected: %v, got: %v", withdrawingVal.Sequence()+1, trx.Sequence())
	}
	if withdrawingVal.Stake() < pld.Amount+trx.Fee() {
		return errors.Errorf(errors.ErrInsufficientFunds, "Insufficient balance")
	}
	if withdrawingVal.UnbondingHeight() == 0 {
		return errors.Errorf(errors.ErrInvalidTx, "Need to unbond first")
	}
	if sb.CurrentHeight() < withdrawingVal.UnbondingHeight()+sb.UnbondInterval() {
		return errors.Errorf(errors.ErrInvalidTx, "Hasn't passed unbonding period, Expected: %v, got: %v", withdrawingVal.UnbondingHeight()+sb.UnbondInterval(), sb.CurrentHeight())
	}

	depositAcc := sb.Account(pld.To)
	if depositAcc == nil {
		depositAcc = sb.MakeNewAccount(pld.To)
	}

	withdrawingVal.IncSequence()
	withdrawingVal.AddToStake(-1 * (pld.Amount + trx.Fee()))
	depositAcc.AddToBalance(pld.Amount)

	sb.UpdateValidator(withdrawingVal)
	sb.UpdateAccount(depositAcc)

	e.fee = trx.Fee()

	return nil
}

func (e *WithdrawExecutor) Fee() int64 {
	return e.fee
}
