package steps

import (
	"context"

	cprocessor "github.com/imfact-labs/currency-model/operation/processor"
	ctype "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/dao-model/operation/dao"
	"github.com/imfact-labs/dao-model/runtime/contracts"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/ps"
)

var PNameOperationProcessorsMap = ps.Name("mitum-dao-operation-processors-map")

type processorInfoA struct {
	hint      hint.Hint
	processor ctype.GetNewProcessor
}

type processorInfoB struct {
	hint      hint.Hint
	processor ctype.GetNewProcessorWithProposal
}

func POperationProcessorsMap(pctx context.Context) (context.Context, error) {
	var isaacParams *isaac.Params
	var db isaac.Database
	var opr *cprocessor.OperationProcessor
	var setA *hint.CompatibleSet[isaac.NewOperationProcessorInternalFunc]
	var setB *hint.CompatibleSet[contracts.NewOperationProcessorInternalWithProposalFunc]

	if err := util.LoadFromContextOK(pctx,
		launch.ISAACParamsContextKey, &isaacParams,
		launch.CenterDatabaseContextKey, &db,
		contracts.OperationProcessorContextKey, &opr,
		launch.OperationProcessorsMapContextKey, &setA,
		contracts.OperationProcessorsMapBContextKey, &setB,
	); err != nil {
		return pctx, err
	}

	//err := opr.SetCheckDuplicationFunc(processor.CheckDuplication)
	//if err != nil {
	//	return pctx, err
	//}
	err := opr.SetGetNewProcessorFunc(cprocessor.GetNewProcessor)
	if err != nil {
		return pctx, err
	}

	processorsA := []processorInfoA{
		{dao.RegisterModelHint, dao.NewRegisterModelProcessor()},
		{dao.UpdateModelConfigHint, dao.NewUpdatePolicyProcessor()},
		{dao.ProposeHint, dao.NewProposeProcessor()},
	}
	processorsB := []processorInfoB{
		{dao.CancelProposalHint, dao.NewCancelProposalProcessor()},
		{dao.RegisterHint, dao.NewRegisterProcessor()},
		{dao.PreSnapHint, dao.NewPreSnapProcessor()},
		{dao.VoteHint, dao.NewVoteProcessor()},
		{dao.PostSnapHint, dao.NewPostSnapProcessor()},
		{dao.ExecuteHint, dao.NewExecuteProcessor()},
	}

	for i := range processorsA {
		p := processorsA[i]

		if err := opr.SetProcessor(p.hint, p.processor); err != nil {
			return pctx, err
		}

		if err := setA.Add(p.hint,
			func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
				return opr.New(
					height,
					getStatef,
					nil,
					nil,
				)
			},
		); err != nil {
			return pctx, err
		}
	}

	for i := range processorsB {
		p := processorsB[i]

		if err := opr.SetProcessorWithProposal(p.hint, p.processor); err != nil {
			return pctx, err
		}

		if err := setB.Add(p.hint,
			func(height base.Height, proposal base.ProposalSignFact, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
				if err := opr.SetProposal(&proposal); err != nil {
					return nil, err
				}

				return opr.New(
					height,
					getStatef,
					nil,
					nil,
				)
			},
		); err != nil {
			return pctx, err
		}
	}

	pctx = context.WithValue(pctx, contracts.OperationProcessorContextKey, opr)
	pctx = context.WithValue(pctx, launch.OperationProcessorsMapContextKey, setA)     //revive:disable-line:modifies-parameter
	pctx = context.WithValue(pctx, contracts.OperationProcessorsMapBContextKey, setB) //revive:disable-line:modifies-parameter

	return pctx, nil
}
