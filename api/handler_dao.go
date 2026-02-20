package api

import (
	"net/http"

	apic "github.com/imfact-labs/currency-model/api"
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/dao-model/digest"
	"github.com/imfact-labs/dao-model/state"
	"github.com/imfact-labs/dao-model/types"
	mitumutil "github.com/imfact-labs/mitum2/util"
)

var (
	HandlerPathDAOService        = `/dao/{contract:(?i)` + ctypes.REStringAddressString + `}`
	HandlerPathDAOProposal       = `/dao/{contract:(?i)` + ctypes.REStringAddressString + `}/proposal/{proposal_id:` + ctypes.ReSpecialCh + `}`
	HandlerPathDAODelegator      = `/dao/{contract:(?i)` + ctypes.REStringAddressString + `}/proposal/{proposal_id:` + ctypes.ReSpecialCh + `}/registrant/{address:(?i)` + ctypes.REStringAddressString + `}`
	HandlerPathDAOVoters         = `/dao/{contract:(?i)` + ctypes.REStringAddressString + `}/proposal/{proposal_id:` + ctypes.ReSpecialCh + `}/voter`
	HandlerPathDAOVotingPowerBox = `/dao/{contract:(?i)` + ctypes.REStringAddressString + `}/proposal/{proposal_id:` + ctypes.ReSpecialCh + `}/votingpower` // revive:disable-line:line-length-limit
)

func SetHandlers(hd *apic.Handlers) {
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

func HandleDAOService(hd *apic.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := apic.CacheKeyPath(r)
	if err := apic.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := apic.ParseRequest(w, r, "contract")
	if err != nil {
		apic.HTTP2ProblemWithError(w, err, status)

		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleDAODesignInGroup(hd, contract)
	}); err != nil {
		apic.HTTP2HandleError(w, err)
	} else {
		apic.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			apic.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleDAODesignInGroup(hd *apic.Handlers, contract string) (interface{}, error) {
	switch design, err := digest.DAOService(hd.Database(), contract); {
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

func buildDAODesignHal(hd *apic.Handlers, contract string, design types.Design) (apic.Hal, error) {
	h, err := hd.CombineURL(HandlerPathDAOService, "contract", contract)
	if err != nil {
		return nil, err
	}

	hal := apic.NewBaseHal(design, apic.NewHalLink(h, nil))

	return hal, nil
}

func HandleDAOProposal(hd *apic.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := apic.CacheKeyPath(r)
	if err := apic.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := apic.ParseRequest(w, r, "contract")
	if err != nil {
		apic.HTTP2ProblemWithError(w, err, status)

		return
	}

	proposalID, err, status := apic.ParseRequest(w, r, "proposal_id")
	if err != nil {
		apic.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleDAOProposalInGroup(hd, contract, proposalID)
	}); err != nil {
		apic.HTTP2HandleError(w, err)
	} else {
		apic.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			apic.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleDAOProposalInGroup(hd *apic.Handlers, contract, proposalID string) (interface{}, error) {
	switch proposal, err := digest.DAOProposal(hd.Database(), contract, proposalID); {
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

func buildDAOProposalHal(hd *apic.Handlers, contract, proposalID string, proposal state.ProposalStateValue) (apic.Hal, error) {
	h, err := hd.CombineURL(HandlerPathDAOProposal, "contract", contract, "proposal_id", proposalID)
	if err != nil {
		return nil, err
	}

	hal := apic.NewBaseHal(proposal, apic.NewHalLink(h, nil))

	return hal, nil
}

func HandleDAODelegator(hd *apic.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := apic.CacheKeyPath(r)
	if err := apic.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := apic.ParseRequest(w, r, "contract")
	if err != nil {
		apic.HTTP2ProblemWithError(w, err, status)
		return
	}

	proposalID, err, status := apic.ParseRequest(w, r, "proposal_id")
	if err != nil {
		apic.HTTP2ProblemWithError(w, err, status)
		return
	}

	delegator, err, status := apic.ParseRequest(w, r, "address")
	if err != nil {
		apic.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleDAODelegatorInGroup(hd, contract, proposalID, delegator)
	}); err != nil {
		apic.HTTP2HandleError(w, err)
	} else {
		apic.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			apic.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleDAODelegatorInGroup(hd *apic.Handlers, contract, proposalID, delegator string) (interface{}, error) {
	switch delegatorInfo, err := digest.DAODelegatorInfo(hd.Database(), contract, proposalID, delegator); {
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

func buildDAODelegatorHal(hd *apic.Handlers,
	contract, proposalID, delegator string,
	delegatorInfo types.DelegatorInfo,
) (apic.Hal, error) {
	h, err := hd.CombineURL(
		HandlerPathDAODelegator,
		"contract", contract,
		"proposal_id", proposalID,
		"address", delegator,
	)
	if err != nil {
		return nil, err
	}

	hal := apic.NewBaseHal(delegatorInfo, apic.NewHalLink(h, nil))

	return hal, nil
}

func HandleDAOVoters(hd *apic.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := apic.CacheKeyPath(r)
	if err := apic.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := apic.ParseRequest(w, r, "contract")
	if err != nil {
		apic.HTTP2ProblemWithError(w, err, status)
		return
	}

	proposalID, err, status := apic.ParseRequest(w, r, "proposal_id")
	if err != nil {
		apic.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleDAOVotersInGroup(hd, contract, proposalID)
	}); err != nil {
		apic.HTTP2HandleError(w, err)
	} else {
		apic.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			apic.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleDAOVotersInGroup(hd *apic.Handlers, contract, proposalID string) (interface{}, error) {
	switch voters, err := digest.DAOVoters(hd.Database(), contract, proposalID); {
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

func buildDAOVotersHal(hd *apic.Handlers,
	contract, proposalID string, voters []types.VoterInfo,
) (apic.Hal, error) {
	h, err := hd.CombineURL(HandlerPathDAOVoters, "contract", contract, "proposal_id", proposalID)
	if err != nil {
		return nil, err
	}

	hal := apic.NewBaseHal(voters, apic.NewHalLink(h, nil))

	return hal, nil
}

func HandleDAOVotingPowerBox(hd *apic.Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := apic.CacheKeyPath(r)
	if err := apic.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := apic.ParseRequest(w, r, "contract")
	if err != nil {
		apic.HTTP2ProblemWithError(w, err, status)
		return
	}

	proposalID, err, status := apic.ParseRequest(w, r, "proposal_id")
	if err != nil {
		apic.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleDAOVotingPowerBoxInGroup(hd, contract, proposalID)
	}); err != nil {
		apic.HTTP2HandleError(w, err)
	} else {
		apic.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)
		if !shared {
			apic.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleDAOVotingPowerBoxInGroup(hd *apic.Handlers, contract, proposalID string) (interface{}, error) {
	switch votingPowerBox, err := digest.DAOVotingPowerBox(hd.Database(), contract, proposalID); {
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

func buildDAOVotingPowerBoxHal(hd *apic.Handlers,
	contract, proposalID string,
	votingPowerBox types.VotingPowerBox,
) (apic.Hal, error) {
	h, err := hd.CombineURL(
		HandlerPathDAOVotingPowerBox,
		"contract", contract,
		"proposal_id", proposalID,
	)
	if err != nil {
		return nil, err
	}

	hal := apic.NewBaseHal(votingPowerBox, apic.NewHalLink(h, nil))

	return hal, nil
}
