package dao

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extras"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-dao/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

type RegisterModelFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Owner                base.Address       `json:"sender"`
	Contract             base.Address       `json:"contract"`
	Option               types.DAOOption    `json:"option"`
	VotingPowerToken     ctypes.CurrencyID  `json:"voting_power_token"`
	Threshold            common.Big         `json:"threshold"`
	ProposalFee          ctypes.Amount      `json:"proposal_fee"`
	ProposerWhitelist    types.Whitelist    `json:"proposer_whitelist"`
	ProposalReviewPeriod uint64             `json:"proposal_review_period"`
	RegistrationPeriod   uint64             `json:"registration_period"`
	PreSnapshotPeriod    uint64             `json:"pre_snapshot_period"`
	VotingPeriod         uint64             `json:"voting_period"`
	PostSnapshotPeriod   uint64             `json:"post_snapshot_period"`
	ExecutionDelayPeriod uint64             `json:"execution_delay_period"`
	Turnout              types.PercentRatio `json:"turnout"`
	Quorum               types.PercentRatio `json:"quorum"`
	Currency             ctypes.CurrencyID  `json:"currency"`
}

func (fact RegisterModelFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(RegisterModelFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Owner:                 fact.sender,
		Contract:              fact.contract,
		Option:                fact.option,
		VotingPowerToken:      fact.votingPowerToken,
		Threshold:             fact.threshold,
		ProposalFee:           fact.proposalFee,
		ProposerWhitelist:     fact.proposerWhitelist,
		ProposalReviewPeriod:  fact.proposalReviewPeriod,
		RegistrationPeriod:    fact.registrationPeriod,
		PreSnapshotPeriod:     fact.preSnapshotPeriod,
		VotingPeriod:          fact.votingPeriod,
		PostSnapshotPeriod:    fact.postSnapshotPeriod,
		ExecutionDelayPeriod:  fact.executionDelayPeriod,
		Turnout:               fact.turnout,
		Quorum:                fact.quorum,
		Currency:              fact.currency,
	})
}

type RegisterModelFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	Owner                string          `json:"sender"`
	Contract             string          `json:"contract"`
	Option               string          `json:"option"`
	VotingPowerToken     string          `json:"voting_power_token"`
	Threshold            string          `json:"threshold"`
	ProposalFee          json.RawMessage `json:"proposal_fee"`
	ProposerWhitelist    json.RawMessage `json:"proposer_whitelist"`
	ProposalReviewPeriod uint64          `json:"proposal_review_period"`
	RegistrationPeriod   uint64          `json:"registration_period"`
	PreSnapshotPeriod    uint64          `json:"pre_snapshot_period"`
	VotingPeriod         uint64          `json:"voting_period"`
	PostSnapshotPeriod   uint64          `json:"post_snapshot_period"`
	ExecutionDelayPeriod uint64          `json:"execution_delay_period"`
	Turnout              uint            `json:"turnout"`
	Quorum               uint            `json:"quorum"`
	Currency             string          `json:"currency"`
}

func (fact *RegisterModelFact) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var uf RegisterModelFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)
	if err := fact.unpack(enc,
		uf.Owner,
		uf.Contract,
		uf.Option,
		uf.VotingPowerToken,
		uf.Threshold,
		uf.ProposalFee,
		uf.ProposerWhitelist,
		uf.ProposalReviewPeriod,
		uf.RegistrationPeriod,
		uf.PreSnapshotPeriod,
		uf.VotingPeriod,
		uf.PostSnapshotPeriod,
		uf.ExecutionDelayPeriod,
		uf.Turnout,
		uf.Quorum,
		uf.Currency,
	); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	return nil
}

func (op RegisterModel) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(OperationMarshaler{
		BaseOperationJSONMarshaler:           op.BaseOperation.JSONMarshaler(),
		BaseOperationExtensionsJSONMarshaler: op.BaseOperationExtensions.JSONMarshaler(),
	})
}

func (op *RegisterModel) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var ubo common.BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *op)
	}

	op.BaseOperation = ubo

	var ueo extras.BaseOperationExtensions
	if err := ueo.DecodeJSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *op)
	}

	op.BaseOperationExtensions = &ueo

	return nil
}
