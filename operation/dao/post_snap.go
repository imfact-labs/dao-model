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
	PostSnapFactHint = hint.MustNewHint("mitum-dao-post-snap-operation-fact-v0.0.1")
	PostSnapHint     = hint.MustNewHint("mitum-dao-post-snap-operation-v0.0.1")
)

type PostSnapFact struct {
	base.BaseFact
	sender     base.Address
	contract   base.Address
	proposalID string
	currency   types.CurrencyID
}

func NewPostSnapFact(
	token []byte,
	sender base.Address,
	contract base.Address,
	proposalID string,
	currency types.CurrencyID,
) PostSnapFact {
	bf := base.NewBaseFact(PostSnapFactHint, token)
	fact := PostSnapFact{
		BaseFact:   bf,
		sender:     sender,
		contract:   contract,
		proposalID: proposalID,
		currency:   currency,
	}
	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact PostSnapFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact PostSnapFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact PostSnapFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		fact.contract.Bytes(),
		[]byte(fact.proposalID),
		fact.currency.Bytes(),
	)
}

func (fact PostSnapFact) IsValid(b []byte) error {
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
			common.ErrSelfTarget.Wrap(
				errors.Errorf("sender %v is same with contract account", fact.sender)))
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	return nil
}

func (fact PostSnapFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact PostSnapFact) Sender() base.Address {
	return fact.sender
}

func (fact PostSnapFact) Contract() base.Address {
	return fact.contract
}

func (fact PostSnapFact) ProposalID() string {
	return fact.proposalID
}

func (fact PostSnapFact) Currency() types.CurrencyID {
	return fact.currency
}

func (fact PostSnapFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, 2)

	as[0] = fact.sender
	as[1] = fact.contract

	return as, nil
}

func (fact PostSnapFact) FeeBase() map[types.CurrencyID][]common.Big {
	required := make(map[types.CurrencyID][]common.Big)
	required[fact.Currency()] = []common.Big{common.ZeroBig}

	return required
}

func (fact PostSnapFact) FeePayer() base.Address {
	return fact.sender
}

func (fact PostSnapFact) FeeItemCount() (uint, bool) {
	return extras.ZeroItem, extras.HasNoItem
}

func (fact PostSnapFact) FactUser() base.Address {
	return fact.sender
}

func (fact PostSnapFact) Signer() base.Address {
	return fact.sender
}

func (fact PostSnapFact) ActiveContract() []base.Address {
	return []base.Address{fact.contract}
}

func (fact PostSnapFact) DupKey() (map[types.DuplicationKeyType][]string, error) {
	r := make(map[types.DuplicationKeyType][]string)
	r[extras.DuplicationKeyTypeSender] = []string{fact.sender.String()}
	r[processor.DuplicationTypeDAOContractProposal] = []string{fmt.Sprintf("%s:%s", fact.Contract().String(), fact.ProposalID())}

	return r, nil
}

type PostSnap struct {
	extras.ExtendedOperation
}

func NewPostSnap(fact PostSnapFact) PostSnap {
	return PostSnap{
		ExtendedOperation: extras.NewExtendedOperation(PostSnapHint, fact),
	}
}
