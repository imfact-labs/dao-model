package dao

import (
	"fmt"

	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/operation/extras"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/dao-model/operation/processor"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

var (
	RegisterFactHint = hint.MustNewHint("mitum-dao-register-operation-fact-v0.0.1")
	RegisterHint     = hint.MustNewHint("mitum-dao-register-operation-v0.0.1")
)

type RegisterFact struct {
	base.BaseFact
	sender     base.Address
	contract   base.Address
	proposalID string
	approved   base.Address
	currency   types.CurrencyID
}

func NewRegisterFact(
	token []byte,
	sender base.Address,
	contract base.Address,
	proposalID string,
	approved base.Address,
	currency types.CurrencyID,
) RegisterFact {
	bf := base.NewBaseFact(RegisterFactHint, token)
	fact := RegisterFact{
		BaseFact:   bf,
		sender:     sender,
		contract:   contract,
		proposalID: proposalID,
		approved:   approved,
		currency:   currency,
	}
	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact RegisterFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact RegisterFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact RegisterFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		fact.contract.Bytes(),
		[]byte(fact.proposalID),
		fact.approved.Bytes(),
		fact.currency.Bytes(),
	)
}

func (fact RegisterFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if err := util.CheckIsValiders(nil, false,
		fact.sender,
		fact.contract,
		fact.currency,
		fact.approved,
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
			common.ErrSelfTarget.Wrap(
				errors.Errorf("sender %v is same with contract account", fact.sender)))
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	return nil
}

func (fact RegisterFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact RegisterFact) Sender() base.Address {
	return fact.sender
}

func (fact RegisterFact) Contract() base.Address {
	return fact.contract
}

func (fact RegisterFact) ProposalID() string {
	return fact.proposalID
}

func (fact RegisterFact) Approved() base.Address {
	return fact.approved
}

func (fact RegisterFact) Currency() types.CurrencyID {
	return fact.currency
}

func (fact RegisterFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, 3)

	as[0] = fact.sender
	as[1] = fact.contract
	as[2] = fact.approved

	return as, nil
}

func (fact RegisterFact) FeeBase() map[types.CurrencyID][]common.Big {
	required := make(map[types.CurrencyID][]common.Big)
	required[fact.Currency()] = []common.Big{common.ZeroBig}

	return required
}

func (fact RegisterFact) FeePayer() base.Address {
	return fact.sender
}

func (fact RegisterFact) FeeItemCount() (uint, bool) {
	return extras.ZeroItem, extras.HasNoItem
}

func (fact RegisterFact) FactUser() base.Address {
	return fact.sender
}

func (fact RegisterFact) Signer() base.Address {
	return fact.sender
}

func (fact RegisterFact) ActiveContract() []base.Address {
	return []base.Address{fact.contract}
}

func (fact RegisterFact) DupKey() (map[types.DuplicationKeyType][]string, error) {
	r := make(map[types.DuplicationKeyType][]string)
	r[extras.DuplicationKeyTypeSender] = []string{fact.sender.String()}
	r[processor.DuplicationTypeDAOContractProposal] = []string{fmt.Sprintf("%s:%s", fact.Contract().String(), fact.ProposalID())}

	return r, nil
}

type Register struct {
	extras.ExtendedOperation
}

func NewRegister(fact RegisterFact) Register {
	return Register{
		ExtendedOperation: extras.NewExtendedOperation(RegisterHint, fact),
	}
}
