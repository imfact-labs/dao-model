package dao

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	cstate "github.com/ProtoconNet/mitum-currency/v3/state"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-dao/state"
	"github.com/ProtoconNet/mitum-dao/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var registerProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(RegisterProcessor)
	},
}

func (Register) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type RegisterProcessor struct {
	*base.BaseOperationProcessor
	proposal *base.ProposalSignFact
}

func NewRegisterProcessor() ctypes.GetNewProcessorWithProposal {
	return func(
		height base.Height,
		proposal *base.ProposalSignFact,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new RegisterProcessor")

		nopp := registerProcessorPool.Get()
		opp, ok := nopp.(*RegisterProcessor)
		if !ok {
			return nil, errors.Errorf("expected RegisterProcessor, not %T", nopp)
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

func (opp *RegisterProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(RegisterFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", RegisterFact{}, op.Fact())), nil
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	if err := cstate.CheckExistsState(currency.DesignStateKey(fact.Currency()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMCurrencyNF).Errorf("currency id %q", fact.Currency())), nil
	}

	if _, _, _, cErr := cstate.ExistsCAccount(
		fact.Approved(), "approved", true, false, getStateFunc); cErr != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMCAccountNA).
				Errorf("%v: approved %v is contract account", cErr, fact.Approved())), nil
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

	st, err := cstate.ExistsState(state.StateKeyProposal(fact.Contract(), fact.ProposalID()), "proposal", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMStateNF).Errorf("proposal %q in contract account %v", fact.ProposalID(), fact.Contract())), nil
	}

	p, err := state.StateProposalValue(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMStateValInvalid).Errorf(
				"proposal %q in contract account %v", fact.ProposalID(), fact.Contract())), nil
	}

	if p.Status() == types.Canceled {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
				Errorf("already canceled proposal %q in contract account %v", fact.ProposalID(), fact.Contract())), nil
	}

	switch st, found, err := getStateFunc(state.StateKeyVoters(fact.Contract(), fact.ProposalID())); {
	case err != nil:
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateNF).
				Errorf("voters for proposal %q in contract account %v", fact.ProposalID(), fact.Contract())), nil
	case found:
		voters, err := state.StateVotersValue(st)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.Wrap(common.ErrMStateValInvalid).
					Errorf("voters for proposal %q in contract account %v", fact.ProposalID(), fact.Contract())), nil
		}

		//voter := types.VoterInfo{}

		for _, v := range voters {
			if !fact.Approved().Equal(v.Account()) {
				continue
			}
			for _, d := range v.Delegators() {
				if fact.Sender().Equal(d) {
					return nil, base.NewBaseOperationProcessReasonError(
						common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
							Errorf("sender %v already delegates the account %v",
								fact.Sender(),
								fact.Approved(),
							)), nil
				}
			}
		}
	}

	switch st, found, err := getStateFunc(state.StateKeyDelegators(fact.Contract(), fact.ProposalID())); {
	case err != nil:
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateNF).
				Errorf("delegators for proposal %q in contract account %v", fact.ProposalID(), fact.Contract())), nil
	case found:
		delegators, err := state.StateDelegatorsValue(st)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.Wrap(common.ErrMStateValInvalid).
					Errorf("delegators for proposal %q in contract account %v", fact.ProposalID(), fact.Contract())), nil
		}

		for _, delegator := range delegators {
			if delegator.Account().Equal(fact.Sender()) {
				return nil, base.NewBaseOperationProcessReasonError(
					common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
						Errorf("sender %v has already registered itself as voter for proposal %q in contract account %v",
							fact.Sender(), fact.ProposalID(), fact.Contract())), nil
			}
		}
	}

	return ctx, nil, nil
}

func (opp *RegisterProcessor) Process(
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, _ := op.Fact().(RegisterFact)

	st, err := cstate.ExistsState(state.StateKeyProposal(fact.Contract(), fact.ProposalID()), "proposal", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("proposal state not found, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
	}

	p, err := state.StateProposalValue(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("proposal value not found from state, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
	}

	proposal := *opp.proposal
	nowTime := uint64(proposal.ProposalFact().ProposedAt().Unix())

	period, start, end := types.GetPeriodOfCurrentTime(p.Policy(), p.Proposal(), types.Registration, nowTime)
	if period != types.Registration {
		return nil, base.NewBaseOperationProcessReasonError("current time is not within the Registration period, Registration period; start(%d), end(%d), but now(%d)", start, end, nowTime), nil
	}

	var sts []base.StateMergeValue

	smv, err := cstate.CreateNotExistAccount(fact.Approved(), getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("%w", err), nil
	} else if smv != nil {
		sts = append(sts, smv)
	}

	var voters []types.VoterInfo
	switch st, found, err := getStateFunc(state.StateKeyVoters(fact.Contract(), fact.ProposalID())); {
	case err != nil:
		return nil, base.NewBaseOperationProcessReasonError("failed to find voters state, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
	case found:
		vs, err := state.StateVotersValue(st)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to find voters value from state, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
		}

		for i, info := range vs {
			if info.Account().Equal(fact.Approved()) {
				delegators := info.Delegators()
				delegators = append(delegators, fact.Sender())
				vs[i] = types.NewVoterInfo(fact.Approved(), delegators)

				break
			}

			if i == len(vs)-1 {
				vs = append(vs, types.NewVoterInfo(fact.Approved(), []base.Address{fact.Sender()}))
			}
		}
		voters = vs
	default:
		var vs []types.VoterInfo
		vs = append(vs, types.NewVoterInfo(fact.Approved(), []base.Address{fact.Sender()}))
		voters = vs
	}

	sts = append(sts,
		common.NewBaseStateMergeValue(
			state.StateKeyVoters(fact.Contract(), fact.ProposalID()),
			state.NewVotersStateValue(voters),
			func(height base.Height, st base.State) base.StateValueMerger {
				return state.NewVotersStateValueMerger(height, state.StateKeyVoters(fact.Contract(), fact.ProposalID()), st)
			},
		),
	)

	switch st, found, err := getStateFunc(state.StateKeyDelegators(fact.Contract(), fact.ProposalID())); {
	case err != nil:
		return nil, base.NewBaseOperationProcessReasonError("failed to find delegators state, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
	case found:
		delegators, err := state.StateDelegatorsValue(st)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to find delegators value from state, %s,%q: %w", fact.Contract(), fact.ProposalID(), err), nil
		}

		delegators = append(delegators, types.NewDelegatorInfo(fact.Sender(), fact.Approved()))

		sts = append(sts,
			common.NewBaseStateMergeValue(
				state.StateKeyDelegators(fact.Contract(), fact.ProposalID()),
				state.NewDelegatorsStateValue(delegators),
				func(height base.Height, st base.State) base.StateValueMerger {
					return state.NewDelegatorsStateValueMerger(height, state.StateKeyDelegators(fact.Contract(), fact.ProposalID()), st)
				},
			),
		)
	default:
		sts = append(sts,
			common.NewBaseStateMergeValue(
				state.StateKeyDelegators(fact.Contract(), fact.ProposalID()),
				state.NewDelegatorsStateValue([]types.DelegatorInfo{types.NewDelegatorInfo(fact.Sender(), fact.Approved())}),
				func(height base.Height, st base.State) base.StateValueMerger {
					return state.NewDelegatorsStateValueMerger(height, state.StateKeyDelegators(fact.Contract(), fact.ProposalID()), st)
				},
			),
		)
	}

	return sts, nil, nil
}

func (opp *RegisterProcessor) Close() error {
	opp.proposal = nil
	registerProcessorPool.Put(opp)

	return nil
}
