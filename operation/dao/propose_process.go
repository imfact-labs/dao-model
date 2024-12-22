package dao

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum-dao/types"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	cstate "github.com/ProtoconNet/mitum-currency/v3/state"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-dao/state"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var proposeProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(ProposeProcessor)
	},
}

func (Propose) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type ProposeProcessor struct {
	*base.BaseOperationProcessor
}

func NewProposeProcessor() ctypes.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new ProposeProcessor")

		nopp := proposeProcessorPool.Get()
		opp, ok := nopp.(*ProposeProcessor)
		if !ok {
			return nil, errors.Errorf("expected ProposeProcessor, not %T", nopp)
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

func (opp *ProposeProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(ProposeFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", ProposeFact{}, op.Fact())), nil
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	if found, _ := cstate.CheckNotExistsState(state.StateKeyProposal(fact.Contract(), fact.ProposalID()), getStateFunc); found {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateE).
				Errorf("proposal %q already exists in contract account %v", fact.ProposalID(), fact.Contract())), nil
	}

	required := map[string]common.Big{}

	currencyPolicy, err := cstate.ExistsCurrencyPolicy(fact.Currency(), getStateFunc)
	if err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMCurrencyNF).Errorf("currency id %q", fact.Currency())), nil
	}

	fee, err := currencyPolicy.Feeer().Fee(common.ZeroBig)
	if err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateValInvalid).
				Errorf("get fee of currency id %q", fact.Currency())), nil
	}

	required[fact.currency.String()] = fee

	st, err := cstate.ExistsState(state.StateKeyDesign(fact.Contract()), "design", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateNF).Wrap(common.ErrMServiceNF).
				Errorf("dao design for contract account %v", fact.Contract())), nil
	}

	design, err := state.StateDesignValue(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateValInvalid).Wrap(common.ErrMServiceNF).
				Errorf("dao design for contract account %v", fact.Contract())), nil
	}

	if design.Option() != fact.Proposal().Option() {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
				Errorf("dao option != proposal option, dao(%s) != proposal(%s)", design.Option(), fact.Proposal().Option())), nil
	}

	votingPowerToken := design.Policy().VotingPowerToken()
	threshold := design.Policy().Threshold()
	proposeFee := design.Policy().ProposalFee()
	whitelist := design.Policy().Whitelist()

	if _, found := required[votingPowerToken.String()]; !found {
		required[votingPowerToken.String()] = common.ZeroBig
	}

	if _, found := required[proposeFee.Currency().String()]; !found {
		required[proposeFee.Currency().String()] = common.ZeroBig
	}

	required[votingPowerToken.String()] = required[votingPowerToken.String()].Add(threshold)
	required[proposeFee.Currency().String()] = required[proposeFee.Currency().String()].Add(proposeFee.Big())

	for k, v := range required {
		st, err = cstate.ExistsState(currency.BalanceStateKey(fact.Sender(), ctypes.CurrencyID(k)), "sender balance", getStateFunc)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.Wrap(common.ErrMStateNF).
					Errorf("sender %v balance for currency id %q", fact.Sender(), fact.Currency())), nil
		}

		switch b, err := currency.StateBalanceValue(st); {
		case err != nil:
			return nil, base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.Wrap(common.ErrMStateValInvalid).
					Errorf("sender %v balance for currency id %q", fact.Sender(), fact.Currency())), nil
		case b.Big().Compare(v) < 0:
			return nil, base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
					Errorf("not enough balance of sender %v for currency id %q", fact.Sender(), fact.Currency())), nil
		}
	}

	if whitelist.Active() && !whitelist.IsExist(fact.Sender()) {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMAccountNAth).
				Errorf("sender not in whitelist, %s", fact.Sender())), nil
	}

	return ctx, nil, nil
}

func (opp *ProposeProcessor) Process(
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, _ := op.Fact().(ProposeFact)

	var sts []base.StateMergeValue

	st, err := cstate.ExistsState(state.StateKeyDesign(fact.Contract()), "key of design", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("dao not found, %s: %w", fact.Contract(), err), nil
	}

	design, err := state.StateDesignValue(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("dao value not found, %s: %w", fact.Contract(), err), nil
	}

	proposeFee := design.Policy().ProposalFee()

	sts = append(sts,
		cstate.NewStateMergeValue(
			state.StateKeyProposal(fact.Contract(), fact.ProposalID()),
			state.NewProposalStateValue(types.Proposed, "proposed", fact.Proposal(), design.Policy()),
		),
	)

	st, err = cstate.ExistsState(currency.BalanceStateKey(fact.Sender(), proposeFee.Currency()), "key of sender balance", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("sender balance for propose fee not found, %s, %q: %w", fact.Sender(), proposeFee.Currency(), err), nil
	}
	_, err = currency.StateBalanceValue(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("sender balance value for propose fee not found, %s, %q: %w", fact.Sender(), proposeFee.Currency(), err), nil
	}

	sts = append(sts,
		common.NewBaseStateMergeValue(
			st.Key(),
			currency.NewDeductBalanceStateValue(proposeFee),
			func(height base.Height, st base.State) base.StateValueMerger {
				return currency.NewBalanceStateValueMerger(height, st.Key(), proposeFee.Currency(), st)
			},
		),
	)

	cBalanceKey := currency.BalanceStateKey(fact.Contract(), proposeFee.Currency())
	sts = append(sts,
		common.NewBaseStateMergeValue(
			cBalanceKey,
			currency.NewAddBalanceStateValue(proposeFee),
			func(height base.Height, st base.State) base.StateValueMerger {
				return currency.NewBalanceStateValueMerger(height, cBalanceKey, proposeFee.Currency(), st)
			},
		))

	return sts, nil, nil
}

func (opp *ProposeProcessor) Close() error {
	proposeProcessorPool.Put(opp)

	return nil
}
