package dao

import (
	"encoding/json"

	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/operation/extras"
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/dao-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
)

type ProposeFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Owner      base.Address      `json:"sender"`
	Contract   base.Address      `json:"contract"`
	ProposalID string            `json:"proposal_id"`
	Proposal   types.Proposal    `json:"proposal"`
	Currency   ctypes.CurrencyID `json:"currency"`
}

func (fact ProposeFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(ProposeFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Owner:                 fact.sender,
		Contract:              fact.contract,
		ProposalID:            fact.proposalID,
		Proposal:              fact.proposal,
		Currency:              fact.currency,
	})
}

type ProposeFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	Owner      string          `json:"sender"`
	Contract   string          `json:"contract"`
	ProposalID string          `json:"proposal_id"`
	Proposal   json.RawMessage `json:"proposal"`
	Currency   string          `json:"currency"`
}

func (fact *ProposeFact) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var uf ProposeFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)
	if err := fact.unpack(enc,
		uf.Owner,
		uf.Contract,
		uf.ProposalID,
		uf.Proposal,
		uf.Currency,
	); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	return nil
}

func (op Propose) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(OperationMarshaler{
		BaseOperationJSONMarshaler:           op.BaseOperation.JSONMarshaler(),
		BaseOperationExtensionsJSONMarshaler: op.BaseOperationExtensions.JSONMarshaler(),
	})
}

func (op *Propose) DecodeJSON(b []byte, enc encoder.Encoder) error {
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
