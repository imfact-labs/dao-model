package dao

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extras"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-dao/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

var (
	UpdateModelConfigFactHint = hint.MustNewHint("mitum-dao-update-model-config-operation-fact-v0.0.1")
	UpdateModelConfigHint     = hint.MustNewHint("mitum-dao-update-model-config-operation-v0.0.1")
)

type UpdateModelConfigFact struct {
	base.BaseFact
	sender               base.Address
	contract             base.Address
	option               types.DAOOption
	votingPowerToken     ctypes.CurrencyID
	threshold            common.Big
	proposalFee          ctypes.Amount
	proposerWhitelist    types.Whitelist
	proposalReviewPeriod uint64
	registrationPeriod   uint64
	preSnapshotPeriod    uint64
	votingPeriod         uint64
	postSnapshotPeriod   uint64
	executionDelayPeriod uint64
	turnout              types.PercentRatio
	quorum               types.PercentRatio
	currency             ctypes.CurrencyID
}

func NewUpdateModelConfigFact(
	token []byte,
	sender base.Address,
	contract base.Address,
	option types.DAOOption,
	votingPowerToken ctypes.CurrencyID,
	threshold common.Big,
	fee ctypes.Amount,
	whitelist types.Whitelist,
	proposalReviewPeriod,
	registrationPeriod,
	preSnapshotPeriod,
	votingPeriod,
	postSnapshotPeriod,
	executionDelayPeriod uint64,
	turnout, quorum types.PercentRatio,
	currency ctypes.CurrencyID,
) UpdateModelConfigFact {
	bf := base.NewBaseFact(UpdateModelConfigFactHint, token)
	fact := UpdateModelConfigFact{
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

func (fact UpdateModelConfigFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact UpdateModelConfigFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact UpdateModelConfigFact) Bytes() []byte {
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

func (fact UpdateModelConfigFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if err := util.CheckIsValiders(nil, false,
		fact.sender,
		fact.contract,
		fact.votingPowerToken,
		fact.option,
		fact.proposalFee,
		fact.threshold,
		fact.proposerWhitelist,
		fact.turnout,
		fact.quorum,
		fact.currency,
	); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if fact.sender.Equal(fact.contract) {
		return common.ErrFactInvalid.Wrap(
			common.ErrSelfTarget.Wrap(
				errors.Errorf("sender %v is same with contract account", fact.sender)))
	}

	for i := range fact.proposerWhitelist.Accounts() {
		if fact.proposerWhitelist.Accounts()[i].Equal(fact.contract) {
			return common.ErrFactInvalid.Wrap(
				common.ErrSelfTarget.Wrap(errors.Errorf("whitelist account %v is same with contract account", fact.proposerWhitelist.Accounts()[i])))
		}
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

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	return nil
}

func (fact UpdateModelConfigFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact UpdateModelConfigFact) Sender() base.Address {
	return fact.sender
}

func (fact UpdateModelConfigFact) Contract() base.Address {
	return fact.contract
}

func (fact UpdateModelConfigFact) Option() types.DAOOption {
	return fact.option
}

func (fact UpdateModelConfigFact) VotingPowerToken() ctypes.CurrencyID {
	return fact.votingPowerToken
}

func (fact UpdateModelConfigFact) ProposalFee() ctypes.Amount {
	return fact.proposalFee
}

func (fact UpdateModelConfigFact) Threshold() common.Big {
	return fact.threshold
}

func (fact UpdateModelConfigFact) Whitelist() types.Whitelist {
	return fact.proposerWhitelist
}

func (fact UpdateModelConfigFact) ProposalReviewPeriod() uint64 {
	return fact.proposalReviewPeriod
}

func (fact UpdateModelConfigFact) RegistrationPeriod() uint64 {
	return fact.registrationPeriod
}

func (fact UpdateModelConfigFact) PreSnapshotPeriod() uint64 {
	return fact.preSnapshotPeriod
}

func (fact UpdateModelConfigFact) VotingPeriod() uint64 {
	return fact.votingPeriod
}

func (fact UpdateModelConfigFact) PostSnapshotPeriod() uint64 {
	return fact.postSnapshotPeriod
}

func (fact UpdateModelConfigFact) ExecutionDelayPeriod() uint64 {
	return fact.executionDelayPeriod
}

func (fact UpdateModelConfigFact) Turnout() types.PercentRatio {
	return fact.turnout
}

func (fact UpdateModelConfigFact) Quorum() types.PercentRatio {
	return fact.quorum
}

func (fact UpdateModelConfigFact) Currency() ctypes.CurrencyID {
	return fact.currency
}

func (fact UpdateModelConfigFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, 2+len(fact.proposerWhitelist.Accounts()))

	as[0] = fact.sender
	as[1] = fact.contract

	for i, ac := range fact.proposerWhitelist.Accounts() {
		as[i+2] = ac
	}

	return as, nil
}

func (fact UpdateModelConfigFact) FeeBase() map[ctypes.CurrencyID][]common.Big {
	required := make(map[ctypes.CurrencyID][]common.Big)
	required[fact.Currency()] = []common.Big{common.ZeroBig}

	return required
}

func (fact UpdateModelConfigFact) FeePayer() base.Address {
	return fact.sender
}

func (fact UpdateModelConfigFact) FactUser() base.Address {
	return fact.sender
}

func (fact UpdateModelConfigFact) Signer() base.Address {
	return fact.sender
}

func (fact UpdateModelConfigFact) ActiveContractOwnerHandlerOnly() [][2]base.Address {
	return [][2]base.Address{{fact.contract, fact.sender}}
}

type UpdateModelConfig struct {
	extras.ExtendedOperation
}

func NewUpdateModelConfig(fact UpdateModelConfigFact) UpdateModelConfig {
	return UpdateModelConfig{
		ExtendedOperation: extras.NewExtendedOperation(UpdateModelConfigHint, fact),
	}
}
