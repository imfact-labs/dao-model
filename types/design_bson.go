package types

import (
	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (de Design) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":  de.Hint().String(),
			"option": de.option,
			"policy": de.policy,
		})
}

type DesignBSONUnmarshaler struct {
	Hint   string   `bson:"_hint"`
	Option string   `bson:"option"`
	Policy bson.Raw `bson:"policy"`
}

func (de *Design) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of Design")

	var ud DesignBSONUnmarshaler
	if err := enc.Unmarshal(b, &ud); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(ud.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return de.unpack(enc, ht, ud.Option, ud.Policy)
}
