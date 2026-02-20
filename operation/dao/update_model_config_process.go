package dao

import (
	"context"
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

func NewUpdatePolicyProcessor() ctypes.GetNewProcessor {
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

	if err := cstate.CheckExistsState(currency.DesignStateKey(fact.Currency()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMCurrencyNF).Errorf("currency id %q", fact.Currency())), nil
	}

	whitelist := fact.Whitelist().Accounts()
	for _, white := range whitelist {
		if _, _, _, cErr := cstate.ExistsCAccount(white, "whitelist", true, false, getStateFunc); cErr != nil {
			return ctx, base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.Wrap(common.ErrMCAccountNA).
					Errorf("%v: whitelist %v is contract account", cErr, white)), nil
		}
	}

	if st, err := cstate.ExistsState(state.StateKeyDesign(fact.Contract()), "design", getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMServiceNF).Errorf("dao service state for contract account %v",
				fact.Contract(),
			)), nil
	} else if _, err := state.StateDesignValue(st); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMServiceNF).Errorf("dao service state for contract account %v",
				fact.Contract(),
			)), nil
	}

	if err := cstate.CheckExistsState(currency.DesignStateKey(fact.VotingPowerToken()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMCurrencyNF).Errorf("voting power token %q", fact.VotingPowerToken())), nil
	}

	if err := cstate.CheckExistsState(currency.DesignStateKey(fact.proposalFee.Currency()), getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateNF).
				Errorf("proposal fee currency %q", fact.proposalFee.Currency())), nil
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
		return nil, base.NewBaseOperationProcessReasonError("invalid design, %s: %w", fact.Contract(), err), nil
	}

	var sts []base.StateMergeValue

	whitelist := fact.Whitelist().Accounts()
	for _, white := range whitelist {
		smv, err := cstate.CreateNotExistAccount(white, getStateFunc)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("%w", err), nil
		} else if smv != nil {
			sts = append(sts, smv)
		}
	}

	sts = append(sts, cstate.NewStateMergeValue(
		state.StateKeyDesign(fact.Contract()),
		state.NewDesignStateValue(design),
	))

	return sts, nil, nil
}

func (opp *UpdateModelConfigProcessor) Close() error {
	updateModelConfigProcessorPool.Put(opp)

	return nil
}
