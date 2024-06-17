package dao

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-dao/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

var (
	RegisterModelFactHint = hint.MustNewHint("mitum-dao-register-model-operation-fact-v0.0.1")
	RegisterModelHint     = hint.MustNewHint("mitum-dao-register-model-operation-v0.0.1")
)

type RegisterModelFact struct {
	base.BaseFact
	sender               base.Address
	contract             base.Address
	option               types.DAOOption
	votingPowerToken     currencytypes.CurrencyID
	threshold            common.Big
	proposalFee          currencytypes.Amount
	proposerWhitelist    types.Whitelist
	proposalReviewPeriod uint64
	registrationPeriod   uint64
	preSnapshotPeriod    uint64
	votingPeriod         uint64
	postSnapshotPeriod   uint64
	executionDelayPeriod uint64
	turnout              types.PercentRatio
	quorum               types.PercentRatio
	currency             currencytypes.CurrencyID
}

func NewRegisterModelFact(
	token []byte,
	sender base.Address,
	contract base.Address,
	option types.DAOOption,
	votingPowerToken currencytypes.CurrencyID,
	threshold common.Big,
	fee currencytypes.Amount,
	whitelist types.Whitelist,
	proposalReviewPeriod,
	registrationPeriod,
	preSnapshotPeriod,
	votingPeriod,
	postSnapshotPeriod,
	executionDelayPeriod uint64,
	turnout, quorum types.PercentRatio,
	currency currencytypes.CurrencyID,
) RegisterModelFact {
	bf := base.NewBaseFact(RegisterModelFactHint, token)
	fact := RegisterModelFact{
		BaseFact:             bf,
		sender:               sender,
		contract:             contract,
		option:               option,
		votingPowerToken:     votingPowerToken,
		threshold:            threshold,
		proposalFee:          fee,
		proposerWhitelist:    whitelist,
		proposalReviewPeriod: proposalReviewPeriod,
		registrationPeriod:   registrationPeriod,
		preSnapshotPeriod:    preSnapshotPeriod,
		votingPeriod:         votingPeriod,
		executionDelayPeriod: executionDelayPeriod,
		postSnapshotPeriod:   postSnapshotPeriod,
		turnout:              turnout,
		quorum:               quorum,
		currency:             currency,
	}
	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact RegisterModelFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact RegisterModelFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact RegisterModelFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		fact.contract.Bytes(),
		fact.option.Bytes(),
		fact.votingPowerToken.Bytes(),
		fact.threshold.Bytes(),
		fact.proposalFee.Bytes(),
		fact.proposerWhitelist.Bytes(),
		util.Uint64ToBytes(fact.proposalReviewPeriod),
		util.Uint64ToBytes(fact.registrationPeriod),
		util.Uint64ToBytes(fact.preSnapshotPeriod),
		util.Uint64ToBytes(fact.votingPeriod),
		util.Uint64ToBytes(fact.postSnapshotPeriod),
		util.Uint64ToBytes(fact.executionDelayPeriod),
		fact.turnout.Bytes(),
		fact.quorum.Bytes(),
		fact.currency.Bytes(),
	)
}

func (fact RegisterModelFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if err := util.CheckIsValiders(nil, false,
		fact.sender,
		fact.contract,
		fact.option,
		fact.votingPowerToken,
		fact.proposalFee,
		fact.threshold,
		fact.proposerWhitelist,
		fact.turnout,
		fact.quorum,
		fact.currency,
	); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if !fact.proposalFee.Big().OverNil() {
		return common.ErrFactInvalid.Wrap(
			common.ErrValOOR.Wrap(errors.Errorf("fee amount must be bigger than or equal to zero, got %v", fact.proposalFee.Big())))
	}

	if !fact.threshold.OverZero() {
		return common.ErrFactInvalid.Wrap(
			common.ErrValOOR.Wrap(errors.Errorf("threshold must be bigger than zero, got %v", fact.threshold)))
	}

	if fact.registrationPeriod == 0 {
		return common.ErrFactInvalid.Wrap(
			common.ErrValOOR.Wrap(
				errors.Errorf("registrationPeriod must be bigger than zero, got %v", fact.registrationPeriod)))
	}

	if fact.preSnapshotPeriod == 0 {
		return common.ErrFactInvalid.Wrap(
			common.ErrValOOR.Wrap(
				errors.Errorf("registrationPeriod must be bigger than zero, got %v", fact.preSnapshotPeriod)))
	}

	if fact.votingPeriod == 0 {
		return common.ErrFactInvalid.Wrap(
			common.ErrValOOR.Wrap(
				errors.Errorf("registrationPeriod must be bigger than zero, got %v", fact.votingPeriod)))
	}

	if fact.postSnapshotPeriod == 0 {
		return common.ErrFactInvalid.Wrap(
			common.ErrValOOR.Wrap(
				errors.Errorf("registrationPeriod must be bigger than zero, got %v", fact.postSnapshotPeriod)))
	}

	if fact.sender.Equal(fact.contract) {
		return common.ErrFactInvalid.Wrap(
			common.ErrSelfTarget.Wrap(errors.Errorf("sender %v is same with contract account", fact.sender)))
	}

	for i := range fact.proposerWhitelist.Accounts() {
		if fact.proposerWhitelist.Accounts()[i].Equal(fact.contract) {
			return common.ErrFactInvalid.Wrap(
				common.ErrSelfTarget.Wrap(errors.Errorf("whitelist account %v is same with contract account", fact.proposerWhitelist.Accounts()[i])))
		}
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	return nil
}

func (fact RegisterModelFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact RegisterModelFact) Sender() base.Address {
	return fact.sender
}

func (fact RegisterModelFact) Contract() base.Address {
	return fact.contract
}

func (fact RegisterModelFact) Option() types.DAOOption {
	return fact.option
}

func (fact RegisterModelFact) VotingPowerToken() currencytypes.CurrencyID {
	return fact.votingPowerToken
}

func (fact RegisterModelFact) ProposalFee() currencytypes.Amount {
	return fact.proposalFee
}

func (fact RegisterModelFact) Threshold() common.Big {
	return fact.threshold
}

func (fact RegisterModelFact) Whitelist() types.Whitelist {
	return fact.proposerWhitelist
}

func (fact RegisterModelFact) ProposalReviewPeriod() uint64 {
	return fact.proposalReviewPeriod
}

func (fact RegisterModelFact) RegistrationPeriod() uint64 {
	return fact.registrationPeriod
}

func (fact RegisterModelFact) PreSnapshotPeriod() uint64 {
	return fact.preSnapshotPeriod
}

func (fact RegisterModelFact) VotingPeriod() uint64 {
	return fact.votingPeriod
}

func (fact RegisterModelFact) PostSnapshotPeriod() uint64 {
	return fact.postSnapshotPeriod
}

func (fact RegisterModelFact) ExecutionDelayPeriod() uint64 {
	return fact.executionDelayPeriod
}

func (fact RegisterModelFact) Turnout() types.PercentRatio {
	return fact.turnout
}

func (fact RegisterModelFact) Quorum() types.PercentRatio {
	return fact.quorum
}

func (fact RegisterModelFact) Currency() currencytypes.CurrencyID {
	return fact.currency
}

func (fact RegisterModelFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, 2+len(fact.proposerWhitelist.Accounts()))

	as[0] = fact.sender
	as[1] = fact.contract

	for i, ac := range fact.proposerWhitelist.Accounts() {
		as[i+2] = ac
	}

	return as, nil
}

type RegisterModel struct {
	common.BaseOperation
}

func NewRegisterModel(fact RegisterModelFact) RegisterModel {
	return RegisterModel{BaseOperation: common.NewBaseOperation(RegisterModelHint, fact)}
}

func (op *RegisterModel) HashSign(priv base.Privatekey, networkID base.NetworkID) error {
	err := op.Sign(priv, networkID)
	if err != nil {
		return err
	}
	return nil
}
