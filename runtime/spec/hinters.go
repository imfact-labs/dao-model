package spec

import (
	"github.com/imfact-labs/dao-model/operation/dao"
	"github.com/imfact-labs/dao-model/state"
	"github.com/imfact-labs/dao-model/types"
	"github.com/imfact-labs/mitum2/util/encoder"
)

var AddedHinters = []encoder.DecodeDetail{
	// revive:disable-next-line:line-length-limit
	{Hint: types.BizProposalHint, Instance: types.BizProposal{}},
	{Hint: types.CryptoProposalHint, Instance: types.CryptoProposal{}},
	{Hint: types.DelegatorInfoHint, Instance: types.DelegatorInfo{}},
	{Hint: types.DesignHint, Instance: types.Design{}},
	{Hint: types.GovernanceCalldataHint, Instance: types.GovernanceCallData{}},
	{Hint: types.PolicyHint, Instance: types.Policy{}},
	{Hint: types.TransferCalldataHint, Instance: types.TransferCallData{}},
	{Hint: types.VoterInfoHint, Instance: types.VoterInfo{}},
	{Hint: types.VotingPowerHint, Instance: types.VotingPower{}},
	{Hint: types.VotingPowerBoxHint, Instance: types.VotingPowerBox{}},
	{Hint: types.WhitelistHint, Instance: types.Whitelist{}},

	{Hint: state.DelegatorsStateValueHint, Instance: state.DelegatorsStateValue{}},
	{Hint: state.DesignStateValueHint, Instance: state.DesignStateValue{}},
	{Hint: state.ProposalStateValueHint, Instance: state.ProposalStateValue{}},
	{Hint: state.VotersStateValueHint, Instance: state.VotersStateValue{}},
	{Hint: state.VotingPowerBoxStateValueHint, Instance: state.VotingPowerBoxStateValue{}},

	{Hint: dao.CancelProposalHint, Instance: dao.CancelProposal{}},
	{Hint: dao.RegisterModelHint, Instance: dao.RegisterModel{}},
	{Hint: dao.ExecuteHint, Instance: dao.Execute{}},
	{Hint: dao.PostSnapHint, Instance: dao.PostSnap{}},
	{Hint: dao.PreSnapHint, Instance: dao.PreSnap{}},
	{Hint: dao.ProposeHint, Instance: dao.Propose{}},
	{Hint: dao.RegisterHint, Instance: dao.Register{}},
	{Hint: dao.UpdateModelConfigHint, Instance: dao.UpdateModelConfig{}},
	{Hint: dao.VoteHint, Instance: dao.Vote{}},
}

var AddedSupportedHinters = []encoder.DecodeDetail{
	{Hint: dao.CancelProposalFactHint, Instance: dao.CancelProposalFact{}},
	{Hint: dao.RegisterModelFactHint, Instance: dao.RegisterModelFact{}},
	{Hint: dao.ExecuteFactHint, Instance: dao.ExecuteFact{}},
	{Hint: dao.PostSnapFactHint, Instance: dao.PostSnapFact{}},
	{Hint: dao.PreSnapFactHint, Instance: dao.PreSnapFact{}},
	{Hint: dao.ProposeFactHint, Instance: dao.ProposeFact{}},
	{Hint: dao.RegisterFactHint, Instance: dao.RegisterFact{}},
	{Hint: dao.UpdateModelConfigFactHint, Instance: dao.UpdateModelConfigFact{}},
	{Hint: dao.VoteFactHint, Instance: dao.VoteFact{}},
}
