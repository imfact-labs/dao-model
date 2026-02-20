package dao

import (
	"context"
	"fmt"
	"sync"

	"github.com/imfact-labs/currency-model/common"
	cstate "github.com/imfact-labs/currency-model/state"
	"github.com/imfact-labs/currency-model/state/currency"
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/dao-model/state"
	"github.com/imfact-labs/dao-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/pkg/errors"
)

var preSnapProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(PreSnapProcessor)
	},
}

func (PreSnap) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type PreSnapProcessor struct {
	*base.BaseOperationProcessor
	proposal *base.ProposalSignFact
}

func NewPreSnapProcessor() ctypes.GetNewProcessorWithProposal {
	return func(
		height base.Height,
		proposal *base.ProposalSignFact,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new PreSnapProcessor")

		nopp := preSnapProcessorPool.Get()
		opp, ok := nopp.(*PreSnapProcessor)
		if !ok {
			return nil, errors.Errorf("expected PreSnapProcessor, not %T", nopp)
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

func (opp *PreSnapProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(PreSnapFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", PreSnapFact{}, op.Fact()),
		), nil
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err),
		), nil
	}

	if err := cstate.CheckExistsState(currency.DesignStateKey(fact.Currency()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMCurrencyNF).Errorf("currency id, %v", fact.Currency()),
		), nil
	}

	if st, err := cstate.ExistsState(state.StateKeyDesign(
		fact.Contract()), "design", getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMServiceNF).Errorf("dao service state for contract account, %v",
				fact.Contract(),
			),
		), nil
	} else if _, err := state.StateDesignValue(st); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMServiceNF).Errorf("dao service state for contract aqccount, %v",
				fact.Contract(),
			),
		), nil
	}

	st, err := cstate.ExistsState(
		state.StateKeyProposal(fact.Contract(), fact.ProposalID()), "proposal", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMStateNF).Errorf(
				"proposal, %s,%v: %v", fact.Contract(), fact.ProposalID(), err),
		), nil
	}

	p, err := state.StateProposalValue(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMStateValInvalid).Errorf(
				"proposal, %s,%v: %v", fact.Contract(), fact.ProposalID(), err),
		), nil
	}

	if p.Status() == types.Canceled {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
				Errorf("already canceled proposal, %s, %q", fact.Contract(), fact.ProposalID()),
		), nil
	} else if p.Status() == types.PreSnapped {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
				Errorf("already preSnapped, %s, %q", fact.Contract(), fact.ProposalID()),
		), nil
	}

	if err := cstate.CheckExistsState(
		state.StateKeyVoters(fact.Contract(), fact.ProposalID()), getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateNF).
				Errorf("voters, %s, %v: %v", fact.Contract(), fact.ProposalID(), err),
		), nil
	}

	if err := cstate.CheckExistsState(
		state.StateKeyDelegators(fact.Contract(), fact.ProposalID()), getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateNF).
				Errorf("delegators, %s, %v: %v", fact.Contract(), fact.ProposalID(), err),
		), nil
	}

	if found, err := cstate.CheckNotExistsState(
		state.StateKeyVotingPowerBox(fact.Contract(), fact.ProposalID()), getStateFunc); found {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateE).
				Errorf("voting power box state already created, %s, %v: %v",
					fact.Contract(), fact.ProposalID(), err),
		), nil
	}

	return ctx, nil, nil
}

func (opp *PreSnapProcessor) Process(
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, _ := op.Fact().(PreSnapFact)

	st, err := cstate.ExistsState(
		state.StateKeyProposal(fact.Contract(), fact.ProposalID()), "proposal", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"proposal not found, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
	}

	p, err := state.StateProposalValue(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"proposal value not found from state, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
	}

	proposal := *opp.proposal
	nowTime := uint64(proposal.ProposalFact().ProposedAt().Unix())

	period, start, end := types.GetPeriodOfCurrentTime(p.Policy(), p.Proposal(), types.PreSnapshot, nowTime)
	if period != types.PreSnapshot {
		return nil, base.NewBaseOperationProcessReasonError(
			"current time is not within the PreSnapshotPeriod, PreSnapshotPeriod; start(%d), end(%d), but now(%d)",
			start, end, nowTime,
		), nil
	}

	var sts []base.StateMergeValue

	var votingPowerBox types.VotingPowerBox
	switch st, found, err := getStateFunc(state.StateKeyVotingPowerBox(fact.Contract(), fact.ProposalID())); {
	case err != nil:
		return nil, base.NewBaseOperationProcessReasonError(
			"failed to find voting power box state, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
	case found:
		if vb, err := state.StateVotingPowerBoxValue(st); err != nil {
			return nil, base.NewBaseOperationProcessReasonError(
				"failed to find voting power box value from state, %s, %q: %w",
				fact.Contract(), fact.ProposalID(), err), nil
		} else {
			votingPowerBox = vb
		}
	default:
		votingPowerBox = types.NewVotingPowerBox(common.ZeroBig, map[string]types.VotingPower{})
	}

	votingPowerToken := p.Policy().VotingPowerToken()

	switch st, found, err := getStateFunc(state.StateKeyVoters(fact.Contract(), fact.ProposalID())); {
	case err != nil:
		return nil, base.NewBaseOperationProcessReasonError(
			"failed to find voters state, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
	case found:
		voters, err := state.StateVotersValue(st)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError(
				"failed to find voters value from state, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
		}

		total := common.ZeroBig
		votingPowers := map[string]types.VotingPower{}
		for _, info := range voters {
			votingPower := common.ZeroBig

			for _, delegator := range info.Delegators() {
				st, err = cstate.ExistsState(
					currency.BalanceStateKey(delegator, votingPowerToken), "key of balance", getStateFunc)
				if err != nil {
					continue
				}

				b, err := currency.StateBalanceValue(st)
				if err != nil {
					return nil, base.NewBaseOperationProcessReasonError(
						"failed to find balance value of the delegator from state, %s, %q: %w",
						delegator, votingPowerToken, err), nil
				}

				votingPower = votingPower.Add(b.Big())
			}

			v, found := votingPowers[info.Account().String()]
			if found {
				votingPower = v.Amount().Add(votingPower)
			}

			votingPowers[info.Account().String()] = types.NewVotingPower(info.Account(), votingPower)
		}

		for _, v := range votingPowers {
			total = total.Add(v.Amount())
		}
		votingPowerBox.SetVotingPowers(votingPowers)
		votingPowerBox.SetTotal(total)
	}

	st, err = cstate.ExistsState(currency.DesignStateKey(votingPowerToken),
		"key of currency design", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"failed to find voting power token currency design, %q: %w", votingPowerToken, err), nil
	}

	currencyDesign, err := currency.GetDesignFromState(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"failed to find voting power token currency design value from state, %q: %w", votingPowerToken, err), nil
	}

	actualTurnoutCount := p.Policy().Turnout().Quorum(currencyDesign.TotalSupply())
	if votingPowerBox.Total().Compare(actualTurnoutCount) < 0 {
		reason := fmt.Sprintf("total voting power, %v is less than turnout, %v", votingPowerBox.Total(), actualTurnoutCount)
		sts = append(sts, cstate.NewStateMergeValue(
			state.StateKeyProposal(fact.Contract(), fact.ProposalID()),
			state.NewProposalStateValue(types.Canceled, reason, p.Proposal(), p.Policy()),
		))
	} else {
		reason := fmt.Sprintf("total voting power, %v is greater than turnout, %v", votingPowerBox.Total(), actualTurnoutCount)
		sts = append(sts,
			cstate.NewStateMergeValue(
				state.StateKeyProposal(fact.Contract(), fact.ProposalID()),
				state.NewProposalStateValue(types.PreSnapped, reason, p.Proposal(), p.Policy()),
			),
			cstate.NewStateMergeValue(
				state.StateKeyVotingPowerBox(fact.Contract(), fact.ProposalID()),
				state.NewVotingPowerBoxStateValue(votingPowerBox),
			),
		)
	}

	return sts, nil, nil
}

func (opp *PreSnapProcessor) Close() error {
	opp.proposal = nil
	preSnapProcessorPool.Put(opp)

	return nil
}
