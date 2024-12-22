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

var voteProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(VoteProcessor)
	},
}

func (Vote) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type VoteProcessor struct {
	*base.BaseOperationProcessor
	proposal *base.ProposalSignFact
}

func NewVoteProcessor() ctypes.GetNewProcessorWithProposal {
	return func(
		height base.Height,
		proposal *base.ProposalSignFact,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new VoteProcessor")

		nopp := voteProcessorPool.Get()
		opp, ok := nopp.(*VoteProcessor)
		if !ok {
			return nil, errors.Errorf("expected VoteProcessor, not %T", nopp)
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

func (opp *VoteProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(VoteFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", VoteFact{}, op.Fact())), nil
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

	if err := cstate.CheckExistsState(currency.DesignStateKey(fact.Currency()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMCurrencyNF).Errorf("fee currency id %q", fact.Currency())), nil
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
				Wrap(common.ErrMStateValInvalid).Errorf("proposal %q in contract account %v", fact.ProposalID(), fact.Contract())), nil
	}

	if p.Status() == types.Canceled {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
				Errorf("already canceled proposal %q in contract account %v", fact.ProposalID(), fact.Contract())), nil
	}

	if p.Status() != types.PreSnapped {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
				Errorf("proposal %q in contract account %v is not in pre-snapped status, got %v", fact.ProposalID(), fact.Contract(), p.Status())), nil
	}

	switch st, found, err := getStateFunc(state.StateKeyVoters(fact.Contract(), fact.ProposalID())); {
	case err != nil:
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateNF).
				Errorf("voters for proposal %q in contract account %v", fact.ProposalID(), fact.Contract())), nil
	case !found:
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateNF).
				Errorf("voters for proposal %q in contract account %v", fact.ProposalID(), fact.Contract())), nil
	default:
		voters, err := state.StateVotersValue(st)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.Wrap(common.ErrMStateValInvalid).
					Errorf("voters for proposal %q in contract account %v", fact.ProposalID(), fact.Contract())), nil
		}

		for i, v := range voters {
			if v.Account().Equal(fact.Sender()) {
				break
			}

			if i == len(voters)-1 {
				return nil, base.NewBaseOperationProcessReasonError(
					common.ErrMPreProcess.Wrap(common.ErrMAccountNAth).
						Errorf("sender %v is not registered as voter for proposal %q in contract account %v",
							fact.Sender(), fact.ProposalID(), fact.Contract())), nil
			}
		}
	}

	switch st, found, err := getStateFunc(state.StateKeyVotingPowerBox(fact.Contract(), fact.ProposalID())); {
	case err != nil:
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateNF).
				Errorf("voting power box for proposal %q in contract account %v", fact.ProposalID(), fact.Contract())), nil
	case found:
		vpb, err := state.StateVotingPowerBoxValue(st)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.Wrap(common.ErrMStateValInvalid).
					Errorf("voting power box for proposal %q in contract account %v", fact.ProposalID(), fact.Contract())), nil
		}

		vp, found := vpb.VotingPowers()[fact.Sender().String()]
		if !found {
			return nil, base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
					Errorf("sender %v has no voting power for proposal %q in contract account %v",
						fact.sender, fact.ProposalID(), fact.Contract())), nil
		}

		if vp.Voted() {
			return nil, base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
					Errorf("sender %v already voted for proposal %q in contract account %v",
						fact.sender, fact.ProposalID(), fact.Contract())), nil
		}
	}

	return ctx, nil, nil
}

func (opp *VoteProcessor) Process(
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, _ := op.Fact().(VoteFact)

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

	period, start, end := types.GetPeriodOfCurrentTime(p.Policy(), p.Proposal(), types.Voting, nowTime)
	if period != types.Voting {
		return nil, base.NewBaseOperationProcessReasonError("current time is not within Voting period, Voting period; start(%d), end(%d), but now(%d)", start, end, nowTime), nil
	}

	var sts []base.StateMergeValue

	var votingPowerBox types.VotingPowerBox
	switch st, found, err := getStateFunc(state.StateKeyVotingPowerBox(fact.Contract(), fact.ProposalID())); {
	case err != nil:
		return nil, base.NewBaseOperationProcessReasonError("failed to find voting power box state, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
	case !found:
		return nil, base.NewBaseOperationProcessReasonError("voting power box state not found, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
	default:
		vpb, err := state.StateVotingPowerBoxValue(st)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to find voting power box value from state, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
		}
		votingPowerBox = vpb
	}

	vp, found := votingPowerBox.VotingPowers()[fact.Sender().String()]
	if !found {
		return nil, base.NewBaseOperationProcessReasonError("sender voting power not found, sender(%s), %s, %q", fact.Sender(), fact.Contract(), fact.ProposalID()), nil
	}
	vp.SetVoted(true)
	vp.SetVoteFor(fact.VoteOption())

	vpb := votingPowerBox.VotingPowers()
	vpb[fact.Sender().String()] = vp
	votingPowerBox.SetVotingPowers(vpb)

	result := votingPowerBox.Result()
	if _, found := result[fact.VoteOption()]; found {
		result[fact.VoteOption()] = result[fact.VoteOption()].Add(vp.Amount())
	} else {
		result[fact.VoteOption()] = common.ZeroBig.Add(vp.Amount())
	}
	votingPowerBox.SetResult(result)

	sts = append(sts,
		cstate.NewStateMergeValue(
			state.StateKeyVotingPowerBox(fact.Contract(), fact.ProposalID()),
			state.NewVotingPowerBoxStateValue(votingPowerBox),
		),
	)

	return sts, nil, nil
}

func (opp *VoteProcessor) Close() error {
	opp.proposal = nil
	voteProcessorPool.Put(opp)

	return nil
}
