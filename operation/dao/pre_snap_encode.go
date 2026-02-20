package dao

import (
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
)

func (fact *PreSnapFact) unpack(enc encoder.Encoder,
	sa, ca, pid, cid string,
) error {
	fact.proposalID = pid
	fact.currency = ctypes.CurrencyID(cid)

	switch a, err := base.DecodeAddress(sa, enc); {
	case err != nil:

	default:
		fact.sender = a
	}

	switch a, err := base.DecodeAddress(ca, enc); {
	case err != nil:

	default:
		fact.contract = a
	}

	return nil
}
