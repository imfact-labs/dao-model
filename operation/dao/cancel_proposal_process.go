package dao

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"sync"

	cstate "github.com/ProtoconNet/mitum-currency/v3/state"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-dao/state"
	"github.com/ProtoconNet/mitum-dao/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var cancelProposalProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CancelProposalProcessor)
	},
}

func (CancelProposal) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type CancelProposalProcessor struct {
	*base.BaseOperationProcessor
	proposal *base.ProposalSignFact
}

func NewCancelProposalProcessor() ctypes.GetNewProcessorWithProposal {
	return func(
		height base.Height,
		proposal *base.ProposalSignFact,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new CancelProposalProcessor")

		nopp := cancelProposalProcessorPool.Get()
		opp, ok := nopp.(*CancelProposalProcessor)
		if !ok {
			return nil, errors.Errorf("expected CancelProposalProcessor, not %T", nopp)
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e.Wrap(err)
		}

		opp.BaseOperationProcessor = b
		opp.proposal = proposal

		return opp, nil
	}
}

func (opp *CancelProposalProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(CancelProposalFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", CancelProposalFact{}, op.Fact())), nil
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	if err := cstate.CheckExistsState(currency.DesignStateKey(fact.Currency()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMCurrencyNF).Errorf("currency id, %v", fact.Currency())), nil
	}

	if st, err := cstate.ExistsState(state.StateKeyDesign(fact.Contract()), "design", getStateFunc); err != nil {
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

	st, err := cstate.ExistsState(
		state.StateKeyProposal(fact.Contract(), fact.ProposalID()), "proposal", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateNF).
				Errorf("proposal %q for contract account %v", fact.ProposalID(), fact.Contract())), nil
	}

	p, err := state.StateProposalValue(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateValInvalid).
				Errorf("proposal %q for contract account %v", fact.ProposalID(), fact.Contract())), nil
	}

	if !fact.Sender().Equal(p.Proposal().Proposer()) {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMAccountNAth).
				Errorf("sender %v is not proposer of the proposal %q in contract account %v", fact.Sender(), fact.ProposalID(), fact.Contract())), nil
	}

	if p.Status() == types.Canceled {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
				Errorf("already canceled proposal %q in contract account %v",
					fact.ProposalID(), fact.Contract())), nil
	} else if p.Status() != types.Proposed {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
				Errorf("proposal %q in contract account %v is %q, but can only be canceled at \"proposed\" status",
					fact.ProposalID(), fact.Contract(), p.Status())), nil
	}

	return ctx, nil, nil
}

func (opp *CancelProposalProcessor) Process(
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, _ := op.Fact().(CancelProposalFact)

	st, err := cstate.ExistsState(state.StateKeyProposal(fact.Contract(), fact.ProposalID()), "proposal", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("proposal not found, %s,%v: %w", fact.Contract(), fact.ProposalID(), err), nil
	}

	p, err := state.StateProposalValue(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("proposal value not found from state, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
	}

	proposal := *opp.proposal
	nowTime := uint64(proposal.ProposalFact().ProposedAt().Unix())

	period, start, _ := types.GetPeriodOfCurrentTime(p.Policy(), p.Proposal(), types.Voting, nowTime)
	if !(period == types.PreLifeCycle || period == types.ProposalReview || period == types.Registration) {
		return nil, base.NewBaseOperationProcessReasonError("cancellable period has passed; voting-started(%d), now(%d)", start, nowTime), nil
	}

	var sts []base.StateMergeValue

	sts = append(sts, cstate.NewStateMergeValue(
		state.StateKeyProposal(fact.Contract(), fact.ProposalID()),
		state.NewProposalStateValue(types.Canceled, "cancel operation processed", p.Proposal(), p.Policy()),
	))

	return sts, nil, nil
}

func (opp *CancelProposalProcessor) Close() error {
	opp.proposal = nil
	cancelProposalProcessorPool.Put(opp)

	return nil
}
