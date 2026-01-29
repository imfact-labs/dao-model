package dao

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extras"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

var (
	ExecuteFactHint = hint.MustNewHint("mitum-dao-execute-operation-fact-v0.0.1")
	ExecuteHint     = hint.MustNewHint("mitum-dao-execute-operation-v0.0.1")
)

type ExecuteFact struct {
	base.BaseFact
	sender     base.Address
	contract   base.Address
	proposalID string
	currency   types.CurrencyID
}

func NewExecuteFact(
	token []byte,
	sender base.Address,
	contract base.Address,
	proposalID string,
	currency types.CurrencyID,
) ExecuteFact {
	bf := base.NewBaseFact(ExecuteFactHint, token)
	fact := ExecuteFact{
		BaseFact:   bf,
		sender:     sender,
		contract:   contract,
		proposalID: proposalID,
		currency:   currency,
	}
	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact ExecuteFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact ExecuteFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact ExecuteFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		fact.contract.Bytes(),
		[]byte(fact.proposalID),
		fact.currency.Bytes(),
	)
}

func (fact ExecuteFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if err := util.CheckIsValiders(nil, false,
		fact.sender,
		fact.contract,
		fact.currency,
	); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if len(fact.proposalID) == 0 {
		return common.ErrFactInvalid.Wrap(common.ErrValOOR.Wrap(errors.Errorf("empty proposal ID")))
	}

	if !types.ReValidSpcecialCh.Match([]byte(fact.proposalID)) {
		return common.ErrFactInvalid.Wrap(
			common.ErrValueInvalid.Wrap(
				errors.Errorf("proposal ID %v must match regex `^[^\\s:/?#\\[\\]$@]*$`", fact.proposalID)))
	}

	if fact.sender.Equal(fact.contract) {
		return common.ErrFactInvalid.Wrap(
			common.ErrSelfTarget.Wrap(errors.Errorf("sender %v is same with contract account", fact.sender)))
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	return nil
}

func (fact ExecuteFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact ExecuteFact) Sender() base.Address {
	return fact.sender
}

func (fact ExecuteFact) Contract() base.Address {
	return fact.contract
}

func (fact ExecuteFact) ProposalID() string {
	return fact.proposalID
}

func (fact ExecuteFact) Currency() types.CurrencyID {
	return fact.currency
}

func (fact ExecuteFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, 2)

	as[0] = fact.sender
	as[1] = fact.contract

	return as, nil
}

func (fact ExecuteFact) FeeBase() map[types.CurrencyID][]common.Big {
	required := make(map[types.CurrencyID][]common.Big)
	required[fact.Currency()] = []common.Big{common.ZeroBig}

	return required
}

func (fact ExecuteFact) FeePayer() base.Address {
	return fact.sender
}

func (fact ExecuteFact) FeeItemCount() (uint, bool) {
	return extras.ZeroItem, extras.HasNoItem
}

func (fact ExecuteFact) FactUser() base.Address {
	return fact.sender
}

func (fact ExecuteFact) Signer() base.Address {
	return fact.sender
}

func (fact ExecuteFact) ActiveContract() []base.Address {
	return []base.Address{fact.contract}
}

type Execute struct {
	extras.ExtendedOperation
}

func NewExecute(fact ExecuteFact) Execute {
	return Execute{
		ExtendedOperation: extras.NewExtendedOperation(ExecuteHint, fact),
	}
}
