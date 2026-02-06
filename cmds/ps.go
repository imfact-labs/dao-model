package cmds

import (
	"context"

	ccmds "github.com/ProtoconNet/mitum-currency/v3/cmds"
	cprocessor "github.com/ProtoconNet/mitum-currency/v3/operation/processor"
	"github.com/ProtoconNet/mitum-dao/operation/dao"
	"github.com/ProtoconNet/mitum-dao/operation/processor"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/isaac"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/ps"
)

var PNameOperationProcessorsMap = ps.Name("mitum-dao-operation-processors-map")

func POperationProcessorsMap(pctx context.Context) (context.Context, error) {
	var isaacParams *isaac.Params
	var db isaac.Database
	var opr *cprocessor.OperationProcessor
	var setA *hint.CompatibleSet[isaac.NewOperationProcessorInternalFunc]
	var setB *hint.CompatibleSet[ccmds.NewOperationProcessorInternalWithProposalFunc]

	if err := util.LoadFromContextOK(pctx,
		launch.ISAACParamsContextKey, &isaacParams,
		launch.CenterDatabaseContextKey, &db,
		ccmds.OperationProcessorContextKey, &opr,
		launch.OperationProcessorsMapContextKey, &setA,
		ccmds.OperationProcessorsMapBContextKey, &setB,
	); err != nil {
		return pctx, err
	}

	//err := opr.SetCheckDuplicationFunc(processor.CheckDuplication)
	//if err != nil {
	//	return pctx, err
	//}
	err := opr.SetGetNewProcessorFunc(processor.GetNewProcessor)
	if err != nil {
		return pctx, err
	}

	if err := opr.SetProcessor(
		dao.RegisterModelHint,
		dao.NewRegisterModelProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		dao.UpdateModelConfigHint,
		dao.NewUpdatePolicyProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		dao.ProposeHint,
		dao.NewProposeProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessorWithProposal(
		dao.CancelProposalHint,
		dao.NewCancelProposalProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessorWithProposal(
		dao.RegisterHint,
		dao.NewRegisterProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessorWithProposal(
		dao.PreSnapHint,
		dao.NewPreSnapProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessorWithProposal(
		dao.VoteHint,
		dao.NewVoteProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessorWithProposal(
		dao.PostSnapHint,
		dao.NewPostSnapProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessorWithProposal(
		dao.ExecuteHint,
		dao.NewExecuteProcessor(),
	); err != nil {
		return pctx, err
	}

	_ = setA.Add(dao.RegisterModelHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			return opr.New(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	_ = setA.Add(dao.UpdateModelConfigHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			return opr.New(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	_ = setA.Add(dao.ProposeHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			return opr.New(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	_ = setB.Add(dao.CancelProposalHint,
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
		})

	_ = setB.Add(dao.RegisterHint,
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
		})

	_ = setB.Add(dao.PreSnapHint,
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
		})

	_ = setB.Add(dao.VoteHint,
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
		})

	_ = setB.Add(dao.PostSnapHint,
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
		})

	_ = setB.Add(dao.ExecuteHint,
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
		})

	pctx = context.WithValue(pctx, ccmds.OperationProcessorContextKey, opr)
	pctx = context.WithValue(pctx, launch.OperationProcessorsMapContextKey, setA) //revive:disable-line:modifies-parameter
	pctx = context.WithValue(pctx, ccmds.OperationProcessorsMapBContextKey, setB) //revive:disable-line:modifies-parameter

	return pctx, nil
}
