package types

import (
	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (r DelegatorInfo) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":    r.Hint().String(),
			"account":  r.account,
			"approved": r.delegatee,
		},
	)
}

type DelegatorInfoBSONUnmarshaler struct {
	Hint      string `bson:"_hint"`
	Account   string `bson:"account"`
	Delegatee string `bson:"approved"`
}

func (r *DelegatorInfo) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of DelegatorInfo")

	var u DelegatorInfoBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	r.BaseHinter = hint.NewBaseHinter(ht)

	switch a, err := base.DecodeAddress(u.Account, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		r.account = a
	}

	switch a, err := base.DecodeAddress(u.Delegatee, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		r.delegatee = a
	}

	return nil
}
