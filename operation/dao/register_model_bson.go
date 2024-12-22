package dao

import (
	"github.com/ProtoconNet/mitum-currency/v3/operation/extras"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

func (fact RegisterModelFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":                  fact.Hint().String(),
			"sender":                 fact.sender,
			"contract":               fact.contract,
			"option":                 fact.option,
			"voting_power_token":     fact.votingPowerToken,
			"threshold":              fact.threshold,
			"proposal_fee":           fact.proposalFee,
			"proposer_whitelist":     fact.proposerWhitelist,
			"proposal_review_period": fact.proposalReviewPeriod,
			"registration_period":    fact.registrationPeriod,
			"pre_snapshot_period":    fact.preSnapshotPeriod,
			"voting_period":          fact.votingPeriod,
			"post_snapshot_period":   fact.postSnapshotPeriod,
			"execution_delay_period": fact.executionDelayPeriod,
			"turnout":                fact.turnout,
			"quorum":                 fact.quorum,
			"currency":               fact.currency,
			"hash":                   fact.BaseFact.Hash().String(),
			"token":                  fact.BaseFact.Token(),
		},
	)
}

type RegisterModelFactBSONUnmarshaler struct {
	Hint                 string   `bson:"_hint"`
	Sender               string   `bson:"sender"`
	Contract             string   `bson:"contract"`
	Option               string   `bson:"option"`
	VotingPowerToken     string   `bson:"voting_power_token"`
	Threshold            string   `bson:"threshold"`
	ProposalFee          bson.Raw `bson:"proposal_fee"`
	ProposerWhitelist    bson.Raw `bson:"proposer_whitelist"`
	ProposalReviewPeriod uint64   `bson:"proposal_review_period"`
	RegistrationPeriod   uint64   `bson:"registration_period"`
	PreSnapshotPeriod    uint64   `bson:"pre_snapshot_period"`
	VotingPeriod         uint64   `bson:"voting_period"`
	PostSnapshotPeriod   uint64   `bson:"post_snapshot_period"`
	ExecutionDelayPeriod uint64   `bson:"execution_delay_period"`
	Turnout              uint     `bson:"turnout"`
	Quorum               uint     `bson:"quorum"`
	Currency             string   `bson:"currency"`
}

func (fact *RegisterModelFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubf common.BaseFactBSONUnmarshaler

	if err := enc.Unmarshal(b, &ubf); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	fact.BaseFact.SetHash(valuehash.NewBytesFromString(ubf.Hash))
	fact.BaseFact.SetToken(ubf.Token)

	var uf RegisterModelFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)
	if err := fact.unpack(enc,
		uf.Sender,
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
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	return nil
}

func (op RegisterModel) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": op.Hint().String(),
			"hash":  op.Hash().String(),
			"fact":  op.Fact(),
			"signs": op.Signs(),
		})
}

func (op *RegisterModel) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo common.BaseOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *op)
	}

	op.BaseOperation = ubo

	var ueo extras.BaseOperationExtensions
	if err := ueo.DecodeBSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *op)
	}

	op.BaseOperationExtensions = &ueo

	return nil
}
