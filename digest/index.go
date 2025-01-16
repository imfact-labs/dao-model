package digest

import (
	cdigest "github.com/ProtoconNet/mitum-currency/v3/digest"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var daoServiceIndexModels = []mongo.IndexModel{
	{
		Keys: bson.D{
			bson.E{Key: "contract", Value: 1},
			bson.E{Key: "height", Value: -1}},
		Options: options.Index().
			SetName(cdigest.IndexPrefix + "dao_service_contract_height"),
	},
}

var daoProposalIndexModels = []mongo.IndexModel{
	{
		Keys: bson.D{
			bson.E{Key: "contract", Value: 1},
			bson.E{Key: "proposal_id", Value: 1},
			bson.E{Key: "height", Value: -1}},
		Options: options.Index().
			SetName(cdigest.IndexPrefix + "dao_proposal_contract_height"),
	},
}

var daoDelegatorsIndexModels = []mongo.IndexModel{
	{
		Keys: bson.D{
			bson.E{Key: "contract", Value: 1},
			bson.E{Key: "proposal_id", Value: 1},
			bson.E{Key: "height", Value: -1}},
		Options: options.Index().
			SetName(cdigest.IndexPrefix + "dao_approved_contract_proposalID_height"),
	},
}

var daoVotersIndexModels = []mongo.IndexModel{
	{
		Keys: bson.D{
			bson.E{Key: "contract", Value: 1},
			bson.E{Key: "proposal_id", Value: 1},
			bson.E{Key: "height", Value: -1}},
		Options: options.Index().
			SetName(cdigest.IndexPrefix + "dao_voter_contract_proposalID_height"),
	},
}

var daoVotingPowerBoxIndexModels = []mongo.IndexModel{
	{
		Keys: bson.D{
			bson.E{Key: "contract", Value: 1},
			bson.E{Key: "proposal_id", Value: 1},
			bson.E{Key: "height", Value: -1}},
		Options: options.Index().
			SetName(cdigest.IndexPrefix + "dao_voting_power_contract_proposalID_height"),
	},
}
var DefaultIndexes = cdigest.DefaultIndexes

func init() {
	DefaultIndexes[DefaultColNameDAO] = daoServiceIndexModels
	DefaultIndexes[DefaultColNameDAOProposal] = daoProposalIndexModels
	DefaultIndexes[DefaultColNameDAODelegators] = daoDelegatorsIndexModels
	DefaultIndexes[DefaultColNameDAOVoters] = daoVotersIndexModels
	DefaultIndexes[DefaultColNameDAOVotingPowerBox] = daoVotingPowerBoxIndexModels
}
