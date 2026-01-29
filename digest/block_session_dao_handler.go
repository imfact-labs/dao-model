package digest

import (
	currencydigest "github.com/ProtoconNet/mitum-currency/v3/digest"
	"github.com/ProtoconNet/mitum-dao/state"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"go.mongodb.org/mongo-driver/mongo"
)

func PrepareDAO(bs *currencydigest.BlockSession, st mitumbase.State) (string, []mongo.WriteModel, error) {

	switch {
	case state.IsStateDesignKey(st.Key()):
		j, err := handleDAODesignState(bs, st)
		if err != nil {
			return "", nil, nil
		}
		return DefaultColNameDAO, j, nil
	case state.IsStateProposalKey(st.Key()):
		j, err := handleDAOProposalState(bs, st)
		if err != nil {
			return "", nil, nil
		}
		return DefaultColNameDAOProposal, j, nil
	case state.IsStateDelegatorsKey(st.Key()):
		j, err := handleDAODelegatorsState(bs, st)
		if err != nil {
			return "", nil, nil
		}
		return DefaultColNameDAODelegators, j, nil
	case state.IsStateVotersKey(st.Key()):
		j, err := handleDAOVotersState(bs, st)
		if err != nil {
			return "", nil, nil
		}

		return DefaultColNameDAOVoters, j, nil
	case state.IsStateVotingPowerBoxKey(st.Key()):
		j, err := handleDAOVotingPowerBoxState(bs, st)
		if err != nil {
			return "", nil, nil
		}

		return DefaultColNameDAOVotingPowerBox, j, nil
	}

	return "", nil, nil
}

func handleDAODesignState(bs *currencydigest.BlockSession, st mitumbase.State) ([]mongo.WriteModel, error) {
	if designDoc, err := NewDAODesignDoc(st, bs.Database().Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(designDoc),
		}, nil
	}
}

func handleDAOProposalState(bs *currencydigest.BlockSession, st mitumbase.State) ([]mongo.WriteModel, error) {
	if nftCollectionDoc, err := NewDAOProposalDoc(st, bs.Database().Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(nftCollectionDoc),
		}, nil
	}
}

func handleDAODelegatorsState(bs *currencydigest.BlockSession, st mitumbase.State) ([]mongo.WriteModel, error) {
	if delegatorsDoc, err := NewDAODelegatorsDoc(st, bs.Database().Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(delegatorsDoc),
		}, nil
	}
}

func handleDAOVotersState(bs *currencydigest.BlockSession, st mitumbase.State) ([]mongo.WriteModel, error) {
	if votersDoc, err := NewDAOVotersDoc(st, bs.Database().Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(votersDoc),
		}, nil
	}
}

func handleDAOVotingPowerBoxState(bs *currencydigest.BlockSession, st mitumbase.State) ([]mongo.WriteModel, error) {
	if nftLastIndexDoc, err := NewDAOVotingPowerBoxDoc(st, bs.Database().Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(nftLastIndexDoc),
		}, nil
	}
}
