package dao

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"sync"

	"github.com/ProtoconNet/mitum-dao/types"

	currencystate "github.com/ProtoconNet/mitum-currency/v3/state"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	stateextension "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-dao/state"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var updateModelConfigProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(UpdateModelConfigProcessor)
	},
}

func (UpdateModelConfig) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type UpdateModelConfigProcessor struct {
	*base.BaseOperationProcessor
}

func NewUpdatePolicyProcessor() currencytypes.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new UpdatePolicyProcessor")

		nopp := updateModelConfigProcessorPool.Get()
		opp, ok := nopp.(*UpdateModelConfigProcessor)
		if !ok {
			return nil, errors.Errorf("expected UpdatePolicyProcessor, not %T", nopp)
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e.Wrap(err)
		}

		opp.BaseOperationProcessor = b

		return opp, nil
	}
}

func (opp *UpdateModelConfigProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(UpdateModelConfigFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", UpdateModelConfigFact{}, op.Fact())), nil
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	if err := currencystate.CheckExistsState(currency.DesignStateKey(fact.Currency()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMCurrencyNF).Errorf("currency id %q", fact.Currency())), nil
	}

	if _, _, aErr, cErr := currencystate.ExistsCAccount(fact.Sender(), "sender", true, false, getStateFunc); aErr != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", aErr)), nil
	} else if cErr != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMCAccountNA).
				Errorf("%v", cErr)), nil
	}

	_, cSt, aErr, cErr := currencystate.ExistsCAccount(fact.Contract(), "contract", true, true, getStateFunc)
	if aErr != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", aErr)), nil
	} else if cErr != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", cErr)), nil
	}

	whitelist := fact.Whitelist().Accounts()
	for _, white := range whitelist {
		if _, _, _, cErr := currencystate.ExistsCAccount(white, "whitelist", true, false, getStateFunc); cErr != nil {
			return ctx, base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.Wrap(common.ErrMCAccountNA).
					Errorf("%v: whitelist %v is contract account", cErr, white)), nil
		}
	}

	//for i := range fact.Whitelist().Accounts() {
	//	if _, _, aErr, cErr := currencystate.ExistsCAccount(
	//		fact.Whitelist().Accounts()[i], "whitelist", true, false, getStateFunc); aErr != nil {
	//		return ctx, base.NewBaseOperationProcessReasonError(
	//			common.ErrMPreProcess.
	//				Errorf("%v", aErr)), nil
	//	} else if cErr != nil {
	//		return ctx, base.NewBaseOperationProcessReasonError(
	//			common.ErrMPreProcess.Wrap(common.ErrMCAccountNA).
	//				Errorf("%v", cErr)), nil
	//	}
	//}

	_, err := stateextension.CheckCAAuthFromState(cSt, fact.Sender())
	if err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	if st, err := currencystate.ExistsState(state.StateKeyDesign(fact.Contract()), "design", getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMServiceNF).Errorf("dao design for contract account %v",
				fact.Contract(),
			)), nil
	} else if _, err := state.StateDesignValue(st); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMServiceNF).Errorf("dao design for contract account %v",
				fact.Contract(),
			)), nil
	}

	if err := currencystate.CheckExistsState(currency.DesignStateKey(fact.VotingPowerToken()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMCurrencyNF).Errorf("voting power token %q", fact.VotingPowerToken())), nil
	}

	if err := currencystate.CheckExistsState(currency.DesignStateKey(fact.proposalFee.Currency()), getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateNF).
				Errorf("proposal fee currency %q", fact.proposalFee.Currency())), nil
	}

	if err := currencystate.CheckFactSignsByState(fact.Sender(), op.Signs(), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMSignInvalid).
				Errorf("%v", err)), nil
	}

	return ctx, nil, nil
}

func (opp *UpdateModelConfigProcessor) Process(
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	e := util.StringError("failed to process UpdatePolicy")

	fact, ok := op.Fact().(UpdateModelConfigFact)
	if !ok {
		return nil, nil, e.Errorf("expected UpdatePolicyFact, not %T", op.Fact())
	}

	policy := types.NewPolicy(
		fact.votingPowerToken, fact.threshold, fact.proposalFee, fact.proposerWhitelist,
		fact.proposalReviewPeriod, fact.registrationPeriod, fact.preSnapshotPeriod, fact.votingPeriod,
		fact.postSnapshotPeriod, fact.executionDelayPeriod, fact.turnout, fact.quorum,
	)
	if err := policy.IsValid(nil); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("invalid dao policy, %s: %w", fact.Contract(), err), nil
	}

	design := types.NewDesign(fact.option, policy)
	if err := design.IsValid(nil); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("invalid dao design, %s: %w", fact.Contract(), err), nil
	}

	var sts []base.StateMergeValue

	whitelist := fact.Whitelist().Accounts()
	for _, white := range whitelist {
		smv, err := currencystate.CreateNotExistAccount(white, getStateFunc)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("%w", err), nil
		} else if smv != nil {
			sts = append(sts, smv)
		}
	}

	sts = append(sts, currencystate.NewStateMergeValue(
		state.StateKeyDesign(fact.Contract()),
		state.NewDesignStateValue(design),
	))

	{ //calculate operation fee
		currencyPolicy, err := currencystate.ExistsCurrencyPolicy(fact.Currency(), getStateFunc)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("currency not found, %q; %w", fact.Currency(), err), nil
		}

		fee, err := currencyPolicy.Feeer().Fee(common.ZeroBig)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError(
				"failed to check fee of currency, %q; %w",
				fact.Currency(),
				err,
			), nil
		}

		senderBalSt, err := currencystate.ExistsState(
			currency.BalanceStateKey(fact.Sender(), fact.Currency()),
			"key of sender balance",
			getStateFunc,
		)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError(
				"sender balance not found, %q; %w",
				fact.Sender(),
				err,
			), nil
		}

		switch senderBal, err := currency.StateBalanceValue(senderBalSt); {
		case err != nil:
			return nil, base.NewBaseOperationProcessReasonError(
				"failed to get balance value, %q; %w",
				currency.BalanceStateKey(fact.Sender(), fact.Currency()),
				err,
			), nil
		case senderBal.Big().Compare(fee) < 0:
			return nil, base.NewBaseOperationProcessReasonError(
				"not enough balance of sender, %q",
				fact.Sender(),
			), nil
		}

		v, ok := senderBalSt.Value().(currency.BalanceStateValue)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", senderBalSt.Value()), nil
		}

		if currencyPolicy.Feeer().Receiver() != nil {
			if err := currencystate.CheckExistsState(currency.AccountStateKey(currencyPolicy.Feeer().Receiver()), getStateFunc); err != nil {
				return nil, nil, err
			} else if feeRcvrSt, found, err := getStateFunc(currency.BalanceStateKey(currencyPolicy.Feeer().Receiver(), fact.currency)); err != nil {
				return nil, nil, err
			} else if !found {
				return nil, nil, errors.Errorf("feeer receiver %s not found", currencyPolicy.Feeer().Receiver())
			} else if feeRcvrSt.Key() != senderBalSt.Key() {
				r, ok := feeRcvrSt.Value().(currency.BalanceStateValue)
				if !ok {
					return nil, nil, errors.Errorf("expected %T, not %T", currency.BalanceStateValue{}, feeRcvrSt.Value())
				}
				sts = append(sts, common.NewBaseStateMergeValue(
					feeRcvrSt.Key(),
					currency.NewAddBalanceStateValue(r.Amount.WithBig(fee)),
					func(height base.Height, st base.State) base.StateValueMerger {
						return currency.NewBalanceStateValueMerger(height, feeRcvrSt.Key(), fact.currency, st)
					},
				))

				sts = append(sts, common.NewBaseStateMergeValue(
					senderBalSt.Key(),
					currency.NewDeductBalanceStateValue(v.Amount.WithBig(fee)),
					func(height base.Height, st base.State) base.StateValueMerger {
						return currency.NewBalanceStateValueMerger(height, senderBalSt.Key(), fact.currency, st)
					},
				))
			}
		}
	}

	//currencyPolicy, err := currencystate.ExistsCurrencyPolicy(fact.Currency(), getStateFunc)
	//if err != nil {
	//	return nil, base.NewBaseOperationProcessReasonError("currency not found, %q: %w", fact.Currency(), err), nil
	//}
	//
	//fee, err := currencyPolicy.Feeer().Fee(common.ZeroBig)
	//if err != nil {
	//	return nil, base.NewBaseOperationProcessReasonError("failed to check fee of currency, %q: %w", fact.Currency(), err), nil
	//}
	//
	//st, err := currencystate.ExistsState(currency.BalanceStateKey(fact.Sender(), fact.Currency()), "key of sender balance", getStateFunc)
	//if err != nil {
	//	return nil, base.NewBaseOperationProcessReasonError("sender balance not found, %s, %q: %w", fact.Sender(), fact.Currency(), err), nil
	//}
	//sb := currencystate.NewStateMergeValue(st.Key(), st.Value())
	//
	//switch b, err := currency.StateBalanceValue(st); {
	//case err != nil:
	//	return nil, base.NewBaseOperationProcessReasonError("failed to get balance value, %s, %q: %w", fact.Sender(), fact.Currency(), err), nil
	//case b.Big().Compare(fee) < 0:
	//	return nil, base.NewBaseOperationProcessReasonError("not enough balance of sender, %s, %q", fact.Sender(), fact.Currency()), nil
	//}
	//
	//v, ok := sb.Value().(currency.BalanceStateValue)
	//if !ok {
	//	return nil, base.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", sb.Value()), nil
	//}
	//sts[1] = currencystate.NewStateMergeValue(sb.Key(), currency.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(fee))))

	return sts, nil, nil
}

func (opp *UpdateModelConfigProcessor) Close() error {
	updateModelConfigProcessorPool.Put(opp)

	return nil
}
