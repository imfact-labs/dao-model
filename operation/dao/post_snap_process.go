package dao

import (
	"context"
	"fmt"
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

var postSnapProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(PostSnapProcessor)
	},
}

func (PostSnap) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type PostSnapProcessor struct {
	*base.BaseOperationProcessor
	proposal *base.ProposalSignFact
}

func NewPostSnapProcessor() ctypes.GetNewProcessorWithProposal {
	return func(
		height base.Height,
		proposal *base.ProposalSignFact,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new PostSnapProcessor")

		nopp := postSnapProcessorPool.Get()
		opp, ok := nopp.(*PostSnapProcessor)
		if !ok {
			return nil, errors.Errorf("expected PostSnapProcessor, not %T", nopp)
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

func (opp *PostSnapProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(PostSnapFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", PostSnapFact{}, op.Fact())), nil
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

	if st, err := cstate.ExistsState(state.StateKeyDesign(fact.Contract()), "design", getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateNF).
				Wrap(common.ErrMServiceNF).Errorf("dao design for contract account %v",
				fact.Contract(),
			)), nil
	} else if _, err := state.StateDesignValue(st); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateValInvalid).
				Wrap(common.ErrMServiceNF).Errorf("dao design for contract account %v",
				fact.Contract(),
			)), nil
	}

	st, err := cstate.ExistsState(state.StateKeyProposal(fact.Contract(), fact.ProposalID()), "proposal", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMStateNF).Errorf("proposal %q for contract account %v", fact.ProposalID(), fact.Contract())), nil
	}

	p, err := state.StateProposalValue(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMStateValInvalid).Errorf("proposal %q for contract account %v", fact.ProposalID(), fact.Contract())), nil
	}

	if p.Status() == types.Canceled {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
				Errorf("already canceled proposal %q for contract account %v", fact.ProposalID(), fact.Contract())), nil
	} else if p.Status() == types.PostSnapped {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
				Errorf("already post snapped proposal %q for contract account %v", fact.ProposalID(), fact.Contract())), nil
	} else if p.Status() == types.Completed {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
				Errorf("already post snapped proposal %q for contract account %v", fact.ProposalID(), fact.Contract())), nil
	}

	if err := cstate.CheckExistsState(state.StateKeyVoters(fact.Contract(), fact.ProposalID()), getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateNF).
				Errorf("voters for proposal %q in contract account %v", fact.ProposalID(), fact.Contract())), nil
	}

	if err := cstate.CheckExistsState(state.StateKeyVotingPowerBox(fact.Contract(), fact.ProposalID()), getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateNF).
				Errorf("voting power box for proposal %q in contract account %v", fact.ProposalID(), fact.Contract())), nil
	}

	return ctx, nil, nil
}

func (opp *PostSnapProcessor) Process(
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, _ := op.Fact().(PostSnapFact)

	st, err := cstate.ExistsState(state.StateKeyProposal(fact.Contract(), fact.ProposalID()), "proposal", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("proposal not found, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
	}

	p, err := state.StateProposalValue(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("proposal value not found from state, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
	}

	proposal := *opp.proposal
	nowTime := uint64(proposal.ProposalFact().ProposedAt().Unix())

	period, start, end := types.GetPeriodOfCurrentTime(p.Policy(), p.Proposal(), types.PostSnapshot, nowTime)
	if period != types.PostSnapshot {
		return nil, base.NewBaseOperationProcessReasonError("current time is not within the PostSnapshotPeriod, PostSnapshotPeriod; start(%d), end(%d), but now(%d)", start, end, nowTime), nil
	}

	var sts []base.StateMergeValue

	if p.Status() != types.PreSnapped {
		sts = append(sts,
			cstate.NewStateMergeValue(
				st.Key(),
				state.NewProposalStateValue(types.Canceled, "post-snap failed as the pre-snap was not executed", p.Proposal(), p.Policy()),
			),
		)

		return sts, nil, nil
	}

	var ovpb types.VotingPowerBox
	switch st, found, err := getStateFunc(state.StateKeyVotingPowerBox(fact.Contract(), fact.ProposalID())); {
	case err != nil:
		return nil, base.NewBaseOperationProcessReasonError("failed to find voting power box state, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
	case found:
		if vb, err := state.StateVotingPowerBoxValue(st); err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to find voting power box value from state, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
		} else {
			ovpb = vb
		}
	default:
		return nil, base.NewBaseOperationProcessReasonError("voting power box state not found, %s, %q", fact.Contract(), fact.ProposalID()), nil
	}

	votingPowerToken := p.Policy().VotingPowerToken()

	var nvpb = types.NewVotingPowerBox(common.ZeroBig, map[string]types.VotingPower{})

	nvps := map[string]types.VotingPower{}
	nvt := common.ZeroBig

	votedTotal := common.ZeroBig
	votingResult := map[uint8]common.Big{}
	// retrieve all voter information for the proposal
	switch st, found, err := getStateFunc(state.StateKeyVoters(fact.Contract(), fact.ProposalID())); {
	case err != nil:
		return nil, base.NewBaseOperationProcessReasonError("failed to find voters state, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
	case found:
		voters, err := state.StateVotersValue(st)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to find voters value, %s, %q: %w", fact.Contract(), fact.ProposalID(), err), nil
		}

		for _, info := range voters {
			a := info.Account().String()
			// if voter did not vote, do not update voting power
			if !ovpb.VotingPowers()[a].Voted() {
				nvps[a] = ovpb.VotingPowers()[a]
				nvt = nvt.Add(nvps[a].Amount())
				continue
			}
			// if voter voted, retrieve all delegated voting power from state
			vp := common.ZeroBig
			for _, delegator := range info.Delegators() {
				st, err = cstate.ExistsState(currency.BalanceStateKey(delegator, votingPowerToken), "key of balance", getStateFunc)
				if err != nil {
					continue
				}

				b, err := currency.StateBalanceValue(st)
				if err != nil {
					return nil, base.NewBaseOperationProcessReasonError("failed to find balance value of the delegator from state, %s, %q: %w", delegator, votingPowerToken, err), nil
				}

				vp = vp.Add(b.Big())
			}
			// compare registered voting power with current voting power, then use the smaller of the two.
			ovp := ovpb.VotingPowers()[a]
			if ovp.Amount().Compare(vp) < 0 {
				nvps[a] = ovp
			} else {
				nvp := types.NewVotingPower(info.Account(), vp)
				nvp.SetVoted(ovp.Voted())
				nvp.SetVoteFor(ovp.VoteFor())

				nvps[a] = nvp
			}
			// count only the voting power of participated voter
			nvt = nvt.Add(nvps[a].Amount())
			// count voting result
			if nvps[a].Voted() {
				if _, found := votingResult[nvps[a].VoteFor()]; !found {
					votingResult[nvps[a].VoteFor()] = common.ZeroBig
				}
				votingResult[nvps[a].VoteFor()] = votingResult[nvps[a].VoteFor()].Add(nvps[a].Amount())
				votedTotal = votedTotal.Add(nvps[a].Amount())
			}
		}

		nvpb.SetVotingPowers(nvps)
		nvpb.SetTotal(nvt)
		nvpb.SetResult(votingResult)
	}

	sts = append(sts, cstate.NewStateMergeValue(
		state.StateKeyVotingPowerBox(fact.Contract(), fact.ProposalID()),
		state.NewVotingPowerBoxStateValue(nvpb),
	))

	st, err = cstate.ExistsState(currency.DesignStateKey(votingPowerToken), "key of currency design", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to find voting power token currency state, %q: %w", votingPowerToken, err), nil
	}

	currencyDesign, err := currency.GetDesignFromState(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to find voting power token currency design value from state, %q: %w", votingPowerToken, err), nil
	}
	//calculate turnout from total supply and quorum from total voted
	actualTurnoutCount := p.Policy().Turnout().Quorum(currencyDesign.TotalSupply())
	actualQuorumCount := p.Policy().Quorum().Quorum(votedTotal)

	r := types.Rejected
	var reason string

	switch {
	case votedTotal.Compare(actualTurnoutCount) < 0:
		r = types.Canceled
		reason = fmt.Sprintf("total votes, %v is less than turnout, %v", votedTotal, actualTurnoutCount)
	case nvpb.Total().Compare(actualQuorumCount) < 0:
		reason = fmt.Sprintf("registerd total voting power, %v is less than quorum, %v", nvpb.Total(), actualQuorumCount)
	case p.Proposal().Option() == types.ProposalCrypto:
		vr0, found0 := votingResult[0]
		vr1, found1 := votingResult[1]
		if !found0 {
			r = types.Rejected
			reason = "no approve vote for crypto proposal"
			break
		} else {
			if found1 {
				if (0 < vr0.Compare(actualQuorumCount)) && (0 < vr0.Compare(vr1)) {
					r = types.Completed
					reason = fmt.Sprintf("approve votes, %v is greater than disapproval votes, %v and exceed the quorum, %v for crypto proposal", vr0, vr1, actualQuorumCount)
					break
				}
			} else {
				if 0 < vr0.Compare(actualQuorumCount) {
					r = types.Completed
					reason = fmt.Sprintf("no disapproval votes and approve votes, %v exceed the quorum,n %v for crypto proposal", vr0, actualQuorumCount)
					break
				}
			}
		}
	case p.Proposal().Option() == types.ProposalBiz:
		options := p.Proposal().VoteOptionsCount() - 1

		var count = 0
		var mvp = common.ZeroBig
		var mvpOption = ^uint8(0)
		var i uint8 = 0
		// check if the vote count for any option is bigger than actual quorum count.
		// last vote option means abstention, so last option is excluded from vote counting.
		for ; i < options; i++ {
			if votingResult[i].Compare(actualQuorumCount) >= 0 {
				if mvp.Compare(votingResult[i]) < 0 {
					count = 1
					mvp = votingResult[i]
					mvpOption = i
				} else if mvp.Equal(votingResult[i]) {
					count += 1
				}
			}
		}

		if count == 1 {
			r = types.Completed
			reason = fmt.Sprintf("voting option %v is greater than any other option and voting result, %v exceed the quorum, %v", mvpOption, mvp, actualQuorumCount)
		}
	}

	sts = append(sts, cstate.NewStateMergeValue(
		state.StateKeyProposal(fact.Contract(), fact.ProposalID()),
		state.NewProposalStateValue(r, reason, p.Proposal(), p.Policy()),
	))

	return sts, nil, nil
}

func (opp *PostSnapProcessor) Close() error {
	opp.proposal = nil
	postSnapProcessorPool.Put(opp)

	return nil
}
