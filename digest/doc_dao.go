package digest

import (
	mongodbst "github.com/ProtoconNet/mitum-currency/v3/digest/mongodb"
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	statedao "github.com/ProtoconNet/mitum-dao/state"
	"github.com/ProtoconNet/mitum-dao/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

type DAODesignDoc struct {
	mongodbst.BaseDoc
	st base.State
	de types.Design
}

func NewDAODesignDoc(st base.State, enc encoder.Encoder) (DAODesignDoc, error) {
	de, err := statedao.StateDesignValue(st)
	if err != nil {
		return DAODesignDoc{}, err
	}
	b, err := mongodbst.NewBaseDoc(nil, st, enc)
	if err != nil {
		return DAODesignDoc{}, err
	}

	return DAODesignDoc{
		BaseDoc: b,
		st:      st,
		de:      de,
	}, nil
}

func (doc DAODesignDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	parsedKey, err := state.ParseStateKey(doc.st.Key(), statedao.DAOPrefix, 3)
	m["contract"] = parsedKey[1]
	m["height"] = doc.st.Height()
	//m["design"] = doc.de

	return bsonenc.Marshal(m)
}

type DAOProposalDoc struct {
	mongodbst.BaseDoc
	st  base.State
	pr  types.Proposal
	ps  types.ProposalStatus
	prs string
}

func NewDAOProposalDoc(st base.State, enc encoder.Encoder) (DAOProposalDoc, error) {
	pv, err := statedao.StateProposalValue(st)
	if err != nil {
		return DAOProposalDoc{}, err
	}
	b, err := mongodbst.NewBaseDoc(nil, st, enc)
	if err != nil {
		return DAOProposalDoc{}, err
	}

	return DAOProposalDoc{
		BaseDoc: b,
		st:      st,
		pr:      pv.Proposal(),
		ps:      pv.Status(),
		prs:     pv.Reason(),
	}, nil
}

func (doc DAOProposalDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	parsedKey, err := state.ParseStateKey(doc.st.Key(), statedao.DAOPrefix, 4)
	m["contract"] = parsedKey[1]
	m["proposal_id"] = parsedKey[2]
	m["height"] = doc.st.Height()
	m["proposal"] = doc.pr
	m["proposal_status"] = doc.ps
	m["proposal_status_reason"] = doc.prs

	return bsonenc.Marshal(m)
}

type DAODelegatorsDoc struct {
	mongodbst.BaseDoc
	st base.State
	di []types.DelegatorInfo
}

func NewDAODelegatorsDoc(st base.State, enc encoder.Encoder) (DAODelegatorsDoc, error) {
	di, err := statedao.StateDelegatorsValue(st)
	if err != nil {
		return DAODelegatorsDoc{}, err
	}
	b, err := mongodbst.NewBaseDoc(nil, st, enc)
	if err != nil {
		return DAODelegatorsDoc{}, err
	}

	return DAODelegatorsDoc{
		BaseDoc: b,
		st:      st,
		di:      di,
	}, nil
}

func (doc DAODelegatorsDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	parsedKey, err := state.ParseStateKey(doc.st.Key(), statedao.DAOPrefix, 4)
	m["contract"] = parsedKey[1]
	m["proposal_id"] = parsedKey[2]
	m["height"] = doc.st.Height()
	m["approved"] = doc.di

	return bsonenc.Marshal(m)
}

type DAOVotersDoc struct {
	mongodbst.BaseDoc
	st base.State
	vi []types.VoterInfo
}

func NewDAOVotersDoc(st base.State, enc encoder.Encoder) (DAOVotersDoc, error) {
	vi, err := statedao.StateVotersValue(st)
	if err != nil {
		return DAOVotersDoc{}, err
	}
	b, err := mongodbst.NewBaseDoc(nil, st, enc)
	if err != nil {
		return DAOVotersDoc{}, err
	}

	return DAOVotersDoc{
		BaseDoc: b,
		st:      st,
		vi:      vi,
	}, nil
}

func (doc DAOVotersDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	parsedKey, err := state.ParseStateKey(doc.st.Key(), statedao.DAOPrefix, 4)
	m["contract"] = parsedKey[1]
	m["proposal_id"] = parsedKey[2]
	m["height"] = doc.st.Height()
	m["voters"] = doc.vi

	return bsonenc.Marshal(m)
}

type DAOVotingPowerBoxDoc struct {
	mongodbst.BaseDoc
	st  base.State
	vpb types.VotingPowerBox
}

func NewDAOVotingPowerBoxDoc(st base.State, enc encoder.Encoder) (DAOVotingPowerBoxDoc, error) {
	vpb, err := statedao.StateVotingPowerBoxValue(st)
	if err != nil {
		return DAOVotingPowerBoxDoc{}, err
	}
	b, err := mongodbst.NewBaseDoc(nil, st, enc)
	if err != nil {
		return DAOVotingPowerBoxDoc{}, err
	}

	return DAOVotingPowerBoxDoc{
		BaseDoc: b,
		st:      st,
		vpb:     vpb,
	}, nil
}

func (doc DAOVotingPowerBoxDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	parsedKey, err := state.ParseStateKey(doc.st.Key(), statedao.DAOPrefix, 4)
	m["contract"] = parsedKey[1]
	m["proposal_id"] = parsedKey[2]
	m["height"] = doc.st.Height()
	m["voting_power_box"] = doc.vpb

	return bsonenc.Marshal(m)
}
