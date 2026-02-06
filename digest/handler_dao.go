package digest

import (
	"net/http"

	cdigest "github.com/ProtoconNet/mitum-currency/v3/digest"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-dao/state"
	"github.com/ProtoconNet/mitum-dao/types"
	mitumutil "github.com/ProtoconNet/mitum2/util"
)

var (
	HandlerPathDAOService        = `/dao/{contract:(?i)` + ctypes.REStringAddressString + `}`
	HandlerPathDAOProposal       = `/dao/{contract:(?i)` + ctypes.REStringAddressString + `}/proposal/{proposal_id:` + ctypes.ReSpecialCh + `}`
	HandlerPathDAODelegator      = `/dao/{contract:(?i)` + ctypes.REStringAddressString + `}/proposal/{proposal_id:` + ctypes.ReSpecialCh + `}/registrant/{address:(?i)` + ctypes.REStringAddressString + `}`
	HandlerPathDAOVoters         = `/dao/{contract:(?i)` + ctypes.REStringAddressString + `}/proposal/{proposal_id:` + ctypes.ReSpecialCh + `}/voter`
	HandlerPathDAOVotingPowerBox = `/dao/{contract:(?i)` + ctypes.REStringAddressString + `}/proposal/{proposal_id:` + ctypes.ReSpecialCh + `}/votingpower` // revive:disable-line:line-length-limit
)

func SetHandlers(hd *cdigest.Handlers) {
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

func HandleDAOService(hd *cdigest.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := cdigest.CacheKeyPath(r)
	if err := cdigest.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := cdigest.ParseRequest(w, r, "contract")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)

		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleDAODesignInGroup(hd, contract)
	}); err != nil {
		cdigest.HTTP2HandleError(w, err)
	} else {
		cdigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			cdigest.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleDAODesignInGroup(hd *cdigest.Handlers, contract string) (interface{}, error) {
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

func buildDAODesignHal(hd *cdigest.Handlers, contract string, design types.Design) (cdigest.Hal, error) {
	h, err := hd.CombineURL(HandlerPathDAOService, "contract", contract)
	if err != nil {
		return nil, err
	}

	hal := cdigest.NewBaseHal(design, cdigest.NewHalLink(h, nil))

	return hal, nil
}

func HandleDAOProposal(hd *cdigest.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := cdigest.CacheKeyPath(r)
	if err := cdigest.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := cdigest.ParseRequest(w, r, "contract")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)

		return
	}

	proposalID, err, status := cdigest.ParseRequest(w, r, "proposal_id")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleDAOProposalInGroup(hd, contract, proposalID)
	}); err != nil {
		cdigest.HTTP2HandleError(w, err)
	} else {
		cdigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			cdigest.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleDAOProposalInGroup(hd *cdigest.Handlers, contract, proposalID string) (interface{}, error) {
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

func buildDAOProposalHal(hd *cdigest.Handlers, contract, proposalID string, proposal state.ProposalStateValue) (cdigest.Hal, error) {
	h, err := hd.CombineURL(HandlerPathDAOProposal, "contract", contract, "proposal_id", proposalID)
	if err != nil {
		return nil, err
	}

	hal := cdigest.NewBaseHal(proposal, cdigest.NewHalLink(h, nil))

	return hal, nil
}

func HandleDAODelegator(hd *cdigest.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := cdigest.CacheKeyPath(r)
	if err := cdigest.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := cdigest.ParseRequest(w, r, "contract")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	proposalID, err, status := cdigest.ParseRequest(w, r, "proposal_id")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	delegator, err, status := cdigest.ParseRequest(w, r, "address")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleDAODelegatorInGroup(hd, contract, proposalID, delegator)
	}); err != nil {
		cdigest.HTTP2HandleError(w, err)
	} else {
		cdigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			cdigest.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleDAODelegatorInGroup(hd *cdigest.Handlers, contract, proposalID, delegator string) (interface{}, error) {
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

func buildDAODelegatorHal(hd *cdigest.Handlers,
	contract, proposalID, delegator string,
	delegatorInfo types.DelegatorInfo,
) (cdigest.Hal, error) {
	h, err := hd.CombineURL(
		HandlerPathDAODelegator,
		"contract", contract,
		"proposal_id", proposalID,
		"address", delegator,
	)
	if err != nil {
		return nil, err
	}

	hal := cdigest.NewBaseHal(delegatorInfo, cdigest.NewHalLink(h, nil))

	return hal, nil
}

func HandleDAOVoters(hd *cdigest.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := cdigest.CacheKeyPath(r)
	if err := cdigest.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := cdigest.ParseRequest(w, r, "contract")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	proposalID, err, status := cdigest.ParseRequest(w, r, "proposal_id")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleDAOVotersInGroup(hd, contract, proposalID)
	}); err != nil {
		cdigest.HTTP2HandleError(w, err)
	} else {
		cdigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			cdigest.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleDAOVotersInGroup(hd *cdigest.Handlers, contract, proposalID string) (interface{}, error) {
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

func buildDAOVotersHal(hd *cdigest.Handlers,
	contract, proposalID string, voters []types.VoterInfo,
) (cdigest.Hal, error) {
	h, err := hd.CombineURL(HandlerPathDAOVoters, "contract", contract, "proposal_id", proposalID)
	if err != nil {
		return nil, err
	}

	hal := cdigest.NewBaseHal(voters, cdigest.NewHalLink(h, nil))

	return hal, nil
}

func HandleDAOVotingPowerBox(hd *cdigest.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := cdigest.CacheKeyPath(r)
	if err := cdigest.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := cdigest.ParseRequest(w, r, "contract")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	proposalID, err, status := cdigest.ParseRequest(w, r, "proposal_id")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleDAOVotingPowerBoxInGroup(hd, contract, proposalID)
	}); err != nil {
		cdigest.HTTP2HandleError(w, err)
	} else {
		cdigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			cdigest.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleDAOVotingPowerBoxInGroup(hd *cdigest.Handlers, contract, proposalID string) (interface{}, error) {
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

func buildDAOVotingPowerBoxHal(hd *cdigest.Handlers,
	contract, proposalID string,
	votingPowerBox types.VotingPowerBox,
) (cdigest.Hal, error) {
	h, err := hd.CombineURL(
		HandlerPathDAOVotingPowerBox,
		"contract", contract,
		"proposal_id", proposalID,
	)
	if err != nil {
		return nil, err
	}

	hal := cdigest.NewBaseHal(votingPowerBox, cdigest.NewHalLink(h, nil))

	return hal, nil
}
