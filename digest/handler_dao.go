package digest

import (
	currencydigest "github.com/ProtoconNet/mitum-currency/v3/digest"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-dao/state"
	"github.com/ProtoconNet/mitum-dao/types"
	mitumutil "github.com/ProtoconNet/mitum2/util"
	"net/http"
)

var (
	HandlerPathDAOService        = `/dao/{contract:(?i)` + currencytypes.REStringAddressString + `}`
	HandlerPathDAOProposal       = `/dao/{contract:(?i)` + currencytypes.REStringAddressString + `}/proposal/{proposal_id:` + currencytypes.ReSpecialCh + `}`
	HandlerPathDAODelegator      = `/dao/{contract:(?i)` + currencytypes.REStringAddressString + `}/proposal/{proposal_id:` + currencytypes.ReSpecialCh + `}/registrant/{address:(?i)` + currencytypes.REStringAddressString + `}`
	HandlerPathDAOVoters         = `/dao/{contract:(?i)` + currencytypes.REStringAddressString + `}/proposal/{proposal_id:` + currencytypes.ReSpecialCh + `}/voter`
	HandlerPathDAOVotingPowerBox = `/dao/{contract:(?i)` + currencytypes.REStringAddressString + `}/proposal/{proposal_id:` + currencytypes.ReSpecialCh + `}/votingpower` // revive:disable-line:line-length-limit
)

func SetHandlers(hd *currencydigest.Handlers) {
	get := 1000
	_ = hd.SetHandler(HandlerPathDAOService, HandleDAOService, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathDAOProposal, HandleDAOProposal, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathDAODelegator, HandleDAODelegator, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathDAOVoters, HandleDAOVoters, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathDAOVotingPowerBox, HandleDAOVotingPowerBox, true, get, get).
		Methods(http.MethodOptions, "GET")
}

func HandleDAOService(hd *currencydigest.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := currencydigest.CacheKeyPath(r)
	if err := currencydigest.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := currencydigest.ParseRequest(w, r, "contract")
	if err != nil {
		currencydigest.HTTP2ProblemWithError(w, err, status)

		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleDAODesignInGroup(hd, contract)
	}); err != nil {
		currencydigest.HTTP2HandleError(w, err)
	} else {
		currencydigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			currencydigest.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleDAODesignInGroup(hd *currencydigest.Handlers, contract string) (interface{}, error) {
	switch design, err := DAOService(hd.Database(), contract); {
	case err != nil:
		return nil, mitumutil.ErrNotFound.WithMessage(err, "dao service, contract %s", contract)
	case design == nil:
		return nil, mitumutil.ErrNotFound.Errorf("dao service, contract %s", contract)
	default:
		hal, err := buildDAODesignHal(hd, contract, *design)
		if err != nil {
			return nil, err
		}
		return hd.Encoder().Marshal(hal)
	}
}

func buildDAODesignHal(hd *currencydigest.Handlers, contract string, design types.Design) (currencydigest.Hal, error) {
	h, err := hd.CombineURL(HandlerPathDAOService, "contract", contract)
	if err != nil {
		return nil, err
	}

	hal := currencydigest.NewBaseHal(design, currencydigest.NewHalLink(h, nil))

	return hal, nil
}

func HandleDAOProposal(hd *currencydigest.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := currencydigest.CacheKeyPath(r)
	if err := currencydigest.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := currencydigest.ParseRequest(w, r, "contract")
	if err != nil {
		currencydigest.HTTP2ProblemWithError(w, err, status)

		return
	}

	proposalID, err, status := currencydigest.ParseRequest(w, r, "proposal_id")
	if err != nil {
		currencydigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleDAOProposalInGroup(hd, contract, proposalID)
	}); err != nil {
		currencydigest.HTTP2HandleError(w, err)
	} else {
		currencydigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			currencydigest.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleDAOProposalInGroup(hd *currencydigest.Handlers, contract, proposalID string) (interface{}, error) {
	switch proposal, err := DAOProposal(hd.Database(), contract, proposalID); {
	case err != nil:
		return nil, mitumutil.ErrNotFound.WithMessage(err, "proposal, contract %s, proposalID %s", contract, proposalID)
	case proposal == nil:
		return nil, mitumutil.ErrNotFound.Errorf("proposal, contract %s, proposalID %s", contract, proposalID)
	default:
		hal, err := buildDAOProposalHal(hd, contract, proposalID, *proposal)
		if err != nil {
			return nil, err
		}
		return hd.Encoder().Marshal(hal)
	}
}

func buildDAOProposalHal(hd *currencydigest.Handlers, contract, proposalID string, proposal state.ProposalStateValue) (currencydigest.Hal, error) {
	h, err := hd.CombineURL(HandlerPathDAOProposal, "contract", contract, "proposal_id", proposalID)
	if err != nil {
		return nil, err
	}

	hal := currencydigest.NewBaseHal(proposal, currencydigest.NewHalLink(h, nil))

	return hal, nil
}

func HandleDAODelegator(hd *currencydigest.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := currencydigest.CacheKeyPath(r)
	if err := currencydigest.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := currencydigest.ParseRequest(w, r, "contract")
	if err != nil {
		currencydigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	proposalID, err, status := currencydigest.ParseRequest(w, r, "proposal_id")
	if err != nil {
		currencydigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	delegator, err, status := currencydigest.ParseRequest(w, r, "address")
	if err != nil {
		currencydigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleDAODelegatorInGroup(hd, contract, proposalID, delegator)
	}); err != nil {
		currencydigest.HTTP2HandleError(w, err)
	} else {
		currencydigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			currencydigest.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleDAODelegatorInGroup(hd *currencydigest.Handlers, contract, proposalID, delegator string) (interface{}, error) {
	switch delegatorInfo, err := DAODelegatorInfo(hd.Database(), contract, proposalID, delegator); {
	case err != nil:
		return nil, mitumutil.ErrNotFound.WithMessage(err, "delegator info, contract %s, proposalID %s, delegator %s", contract, proposalID, delegator)
	case delegatorInfo == nil:
		return nil, mitumutil.ErrNotFound.Errorf("delegator info, contract %s, proposalID %s, delegator %s", contract, proposalID, delegator)
	default:
		hal, err := buildDAODelegatorHal(hd, contract, proposalID, delegator, *delegatorInfo)
		if err != nil {
			return nil, err
		}
		return hd.Encoder().Marshal(hal)
	}
}

func buildDAODelegatorHal(hd *currencydigest.Handlers,
	contract, proposalID, delegator string,
	delegatorInfo types.DelegatorInfo,
) (currencydigest.Hal, error) {
	h, err := hd.CombineURL(
		HandlerPathDAODelegator,
		"contract", contract,
		"proposal_id", proposalID,
		"address", delegator,
	)
	if err != nil {
		return nil, err
	}

	hal := currencydigest.NewBaseHal(delegatorInfo, currencydigest.NewHalLink(h, nil))

	return hal, nil
}

func HandleDAOVoters(hd *currencydigest.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := currencydigest.CacheKeyPath(r)
	if err := currencydigest.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := currencydigest.ParseRequest(w, r, "contract")
	if err != nil {
		currencydigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	proposalID, err, status := currencydigest.ParseRequest(w, r, "proposal_id")
	if err != nil {
		currencydigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleDAOVotersInGroup(hd, contract, proposalID)
	}); err != nil {
		currencydigest.HTTP2HandleError(w, err)
	} else {
		currencydigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			currencydigest.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleDAOVotersInGroup(hd *currencydigest.Handlers, contract, proposalID string) (interface{}, error) {
	switch voters, err := DAOVoters(hd.Database(), contract, proposalID); {
	case err != nil:
		return nil, mitumutil.ErrNotFound.WithMessage(err, "voters, contract %s, proposalID %s", contract, proposalID)
	case voters == nil:
		return nil, mitumutil.ErrNotFound.Errorf("voters, contract %s, proposalID %s", contract, proposalID)
	default:
		hal, err := buildDAOVotersHal(hd, contract, proposalID, voters)
		if err != nil {
			return nil, err
		}
		return hd.Encoder().Marshal(hal)
	}
}

func buildDAOVotersHal(hd *currencydigest.Handlers,
	contract, proposalID string, voters []types.VoterInfo,
) (currencydigest.Hal, error) {
	h, err := hd.CombineURL(HandlerPathDAOVoters, "contract", contract, "proposal_id", proposalID)
	if err != nil {
		return nil, err
	}

	hal := currencydigest.NewBaseHal(voters, currencydigest.NewHalLink(h, nil))

	return hal, nil
}

func HandleDAOVotingPowerBox(hd *currencydigest.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := currencydigest.CacheKeyPath(r)
	if err := currencydigest.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := currencydigest.ParseRequest(w, r, "contract")
	if err != nil {
		currencydigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	proposalID, err, status := currencydigest.ParseRequest(w, r, "proposal_id")
	if err != nil {
		currencydigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleDAOVotingPowerBoxInGroup(hd, contract, proposalID)
	}); err != nil {
		currencydigest.HTTP2HandleError(w, err)
	} else {
		currencydigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			currencydigest.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleDAOVotingPowerBoxInGroup(hd *currencydigest.Handlers, contract, proposalID string) (interface{}, error) {
	switch votingPowerBox, err := DAOVotingPowerBox(hd.Database(), contract, proposalID); {
	case err != nil:
		return nil, mitumutil.ErrNotFound.WithMessage(err, "voting power box, contract %s, proposalID %s", contract, proposalID)
	case votingPowerBox == nil:
		return nil, mitumutil.ErrNotFound.Errorf("voting power box, contract %s, proposalID %s", contract, proposalID)

	default:
		hal, err := buildDAOVotingPowerBoxHal(hd, contract, proposalID, *votingPowerBox)
		if err != nil {
			return nil, err
		}
		return hd.Encoder().Marshal(hal)
	}
}

func buildDAOVotingPowerBoxHal(hd *currencydigest.Handlers,
	contract, proposalID string,
	votingPowerBox types.VotingPowerBox,
) (currencydigest.Hal, error) {
	h, err := hd.CombineURL(
		HandlerPathDAOVotingPowerBox,
		"contract", contract,
		"proposal_id", proposalID,
	)
	if err != nil {
		return nil, err
	}

	hal := currencydigest.NewBaseHal(votingPowerBox, currencydigest.NewHalLink(h, nil))

	return hal, nil
}
