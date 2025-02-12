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

var executeProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(ExecuteProcessor)
	},
}

func (Execute) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type ExecuteProcessor struct {
	*base.BaseOperationProcessor
	proposal *base.ProposalSignFact
}

func NewExecuteProcessor() ctypes.GetNewProcessorWithProposal {
	return func(
		height base.Height,
		proposal *base.ProposalSignFact,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new ExecuteProcessor")

		nopp := executeProcessorPool.Get()
		opp, ok := nopp.(*ExecuteProcessor)
		if !ok {
			return nil, errors.Errorf("expected ExecuteProcessor, not %T", nopp)
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

func (opp *ExecuteProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(ExecuteFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", ExecuteFact{}, op.Fact())), nil
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

	if st, err := cstate.ExistsState(state.StateKeyDesign(
		fact.Contract()), "design", getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMServiceNF).Errorf("dao service state for contract account %v",
				fact.Contract(),
			)), nil
	} else if _, err := state.StateDesignValue(st); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMServiceNF).Errorf("dao service state value for contract account %v",
				fact.Contract(),
			)), nil
	}

	st, err := cstate.ExistsState(
		state.StateKeyProposal(fact.Contract(), fact.ProposalID()), "proposal", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMStateNF).Errorf(
				"proposal state %q for contract account %v", fact.ProposalID(), fact.Contract())), nil
	}

	p, err := state.StateProposalValue(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMStateValInvalid).Errorf(
				"proposal state value %q for contract account %v", fact.ProposalID(), fact.Contract())), nil
	}

	if p.Status() == types.Canceled {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
				Errorf("already canceled proposal %q for contract account %v",
					fact.ProposalID(), fact.Contract())), nil
	} else if p.Status() == types.Rejected {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
				Errorf("already rejected proposal %q for contract account %v",
					fact.ProposalID(), fact.Contract())), nil
	} else if p.Status() == types.Executed {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).
				Errorf("already executed proposal %q for contract account %v",
					fact.ProposalID(), fact.Contract())), nil
	}

	if err := cstate.CheckExistsState(state.StateKeyVotingPowerBox(
		fact.Contract(), fact.ProposalID()), getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateNF).
				Errorf("voting power box of proposal %q in contract account %v",
					fact.ProposalID(), fact.Contract()),
		), nil
	}

	return ctx, nil, nil
}

func (opp *ExecuteProcessor) Process(
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, _ := op.Fact().(ExecuteFact)

	st, err := cstate.ExistsState(state.StateKeyProposal(
		fact.Contract(), fact.ProposalID()), "proposal", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"proposal not found, %s, %q: %w", fact.Contract(), fact.ProposalID(), err,
		), nil
	}

	p, err := state.StateProposalValue(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"proposal value not found from state, %s, %q: %w", fact.Contract(), fact.ProposalID(), err,
		), nil
	}

	proposal := *opp.proposal
	nowTime := uint64(proposal.ProposalFact().ProposedAt().Unix())

	period, start, end := types.GetPeriodOfCurrentTime(p.Policy(), p.Proposal(), types.Execute, nowTime)
	if period != types.Execute {
		return nil, base.NewBaseOperationProcessReasonError(
			"current time is not within the Execution, Execution period; start(%d), end(%d), but now(%d)",
			start, end, nowTime,
		), nil
	}

	var sts []base.StateMergeValue

	if p.Status() != types.Completed {
		sts = append(sts,
			cstate.NewStateMergeValue(
				st.Key(),
				state.NewProposalStateValue(types.Canceled, "execution failed", p.Proposal(), p.Policy()),
			),
		)

		return sts, nil, nil
	}

	sts = append(sts, cstate.NewStateMergeValue(
		state.StateKeyProposal(fact.Contract(), fact.ProposalID()),
		state.NewProposalStateValue(types.Executed, "execution succeeded", p.Proposal(), p.Policy()),
	))

	if p.Proposal().Option() == types.ProposalCrypto {
		cp, _ := p.Proposal().(types.CryptoProposal)

		switch cp.CallData().Type() {
		case types.CalldataTransfer:
			cd, ok := cp.CallData().(types.TransferCallData)
			if !ok {
				return nil, base.NewBaseOperationProcessReasonError(
					"expected TransferCalldata, not %T", cp.CallData()), nil
			}

			if err := cstate.CheckExistsState(currency.AccountStateKey(cd.Sender()), getStateFunc); err != nil {
				return nil, base.NewBaseOperationProcessReasonError(
					"calldata sender not found, %s: %w", cd.Sender(), err), nil
			}

			if err := cstate.CheckExistsState(currency.AccountStateKey(cd.Receiver()), getStateFunc); err != nil {
				return nil, base.NewBaseOperationProcessReasonError(
					"calldata receiver not found, %s: %w", cd.Receiver(), err), nil
			}

			st, err = cstate.ExistsState(
				currency.BalanceStateKey(cd.Sender(), cd.Amount().Currency()), "key of balance", getStateFunc)
			if err != nil {
				return nil, base.NewBaseOperationProcessReasonError(
					"failed to find calldata sender balance, %s, %q: %w", cd.Sender(), cd.Amount().Currency(), err), nil
			}

			sb, err := currency.StateBalanceValue(st)
			if err != nil {
				return nil, base.NewBaseOperationProcessReasonError(
					"failed to find calldata sender balance value, %s, %q: %w", cd.Sender(), cd.Amount().Currency(), err), nil
			}

			if sb.Big().Compare(cd.Amount().Big()) >= 0 {
				sts = append(sts, cstate.NewStateMergeValue(
					st.Key(),
					currency.NewBalanceStateValue(
						ctypes.NewAmount(sb.Big().Sub(cd.Amount().Big()), cd.Amount().Currency()),
					),
				))

				switch st, found, err := getStateFunc(currency.BalanceStateKey(cd.Receiver(), cd.Amount().Currency())); {
				case err != nil:
					return nil, base.NewBaseOperationProcessReasonError(
						"failed to find calldata receiver balance, %s, %q: %w", cd.Receiver(), cd.Amount().Currency(), err), nil
				case found:
					rb, err := currency.StateBalanceValue(st)
					if err != nil {
						return nil, base.NewBaseOperationProcessReasonError(
							"failed to find calldata receiver balance value, %s, %q: %w",
							cd.Receiver(), cd.Amount().Currency(), err), nil
					}

					sts = append(sts, cstate.NewStateMergeValue(
						st.Key(),
						currency.NewBalanceStateValue(
							ctypes.NewAmount(rb.Big().Add(cd.Amount().Big()), cd.Amount().Currency()),
						),
					))
				default:
					sts = append(sts, cstate.NewStateMergeValue(
						st.Key(),
						currency.NewBalanceStateValue(
							ctypes.NewAmount(common.ZeroBig.Add(cd.Amount().Big()), cd.Amount().Currency()),
						),
					))
				}
			}
		case types.CalldataGovernance:
			cd, ok := cp.CallData().(types.GovernanceCallData)
			if !ok {
				return nil, base.NewBaseOperationProcessReasonError(
					"expected GovernanceCalldata, not %T", cp.CallData()), nil
			}

			st, err := cstate.ExistsState(state.StateKeyDesign(fact.Contract()), "key of design", getStateFunc)
			if err != nil {
				return nil, base.NewBaseOperationProcessReasonError(
					common.ErrMServiceNF.Errorf(
						"dao service state for contract account, %v: %v", fact.Contract(), err)), nil
			}

			design, err := state.StateDesignValue(st)
			if err != nil {
				return nil, base.NewBaseOperationProcessReasonError(common.ErrMStateInvalid.Errorf(
					"dao service state value for contract account, %v: %v", fact.Contract(), err)), nil
			}

			nd := types.NewDesign(design.Option(), cd.Policy())

			if err := nd.IsValid(nil); err != nil {
				sts = append(sts, cstate.NewStateMergeValue(
					state.StateKeyDesign(fact.Contract()),
					state.NewDesignStateValue(
						nd,
					),
				))
			}
		default:
			return nil, base.NewBaseOperationProcessReasonError(
				"invalid calldata, %s, %q", fact.Contract(), fact.ProposalID()), nil
		}
	}

	return sts, nil, nil
}

func (opp *ExecuteProcessor) Close() error {
	opp.proposal = nil
	executeProcessorPool.Put(opp)

	return nil
}
