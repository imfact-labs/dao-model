package types

import (
	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (wl Whitelist) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":    wl.Hint().String(),
			"active":   wl.active,
			"accounts": wl.accounts,
		},
	)
}

type WhitelistBSONUnmarshaler struct {
	Hint     string   `bson:"_hint"`
	Active   bool     `bson:"active"`
	Accounts []string `bson:"accounts"`
}

func (wl *Whitelist) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of Whitelist")

	var uw WhitelistBSONUnmarshaler
	if err := enc.Unmarshal(b, &uw); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(uw.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return wl.unpack(enc, ht, uw.Active, uw.Accounts)
}

func (po Policy) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":                  po.Hint().String(),
			"voting_power_token":     po.votingPowerToken,
			"threshold":              po.threshold,
			"proposal_fee":           po.proposalFee,
			"proposer_whitelist":     po.proposerWhitelist,
			"proposal_review_period": po.proposalReviewPeriod,
			"registration_period":    po.registrationPeriod,
			"pre_snapshot_period":    po.preSnapshotPeriod,
			"voting_period":          po.votingPeriod,
			"post_snapshot_period":   po.postSnapshotPeriod,
			"execution_delay_period": po.executionDelayPeriod,
			"turnout":                po.turnout,
			"quorum":                 po.quorum,
		},
	)
}

type PolicyBSONUnmarshaler struct {
	Hint                 string   `bson:"_hint"`
	Token                string   `bson:"voting_power_token"`
	Threshold            string   `bson:"threshold"`
	Fee                  bson.Raw `bson:"proposal_fee"`
	Whitelist            bson.Raw `bson:"proposer_whitelist"`
	ProposalReviewPeriod uint64   `bson:"proposal_review_period"`
	RegistrationPeriod   uint64   `bson:"registration_period"`
	PreSnapshotPeriod    uint64   `bson:"pre_snapshot_period"`
	VotingPeriod         uint64   `bson:"voting_period"`
	PostSnapshotPeriod   uint64   `bson:"post_snapshot_period"`
	ExecutionDelayPeriod uint64   `bson:"execution_delay_period"`
	Turnout              uint     `bson:"turnout"`
	Quorum               uint     `bson:"quorum"`
}

func (po *Policy) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of Policy")

	var upo PolicyBSONUnmarshaler
	if err := enc.Unmarshal(b, &upo); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(upo.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return po.unpack(enc, ht,
		upo.Token,
		upo.Threshold,
		upo.Fee,
		upo.Whitelist,
		upo.ProposalReviewPeriod,
		upo.RegistrationPeriod,
		upo.PreSnapshotPeriod,
		upo.VotingPeriod,
		upo.PostSnapshotPeriod,
		upo.ExecutionDelayPeriod,
		upo.Turnout,
		upo.Quorum,
	)
}
