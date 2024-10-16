package rest

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/google/uuid"
	coreproposal "github.com/goverland-labs/goverland-core-sdk-go/proposal"
	"golang.org/x/exp/slices"

	"github.com/goverland-labs/inbox-web-api/internal/entities/common"

	"github.com/gorilla/mux"
	coresdk "github.com/goverland-labs/goverland-core-sdk-go"
	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/auth"
	internaldao "github.com/goverland-labs/inbox-web-api/internal/entities/dao"
	"github.com/goverland-labs/inbox-web-api/internal/entities/proposal"
	internaltools "github.com/goverland-labs/inbox-web-api/internal/entities/tools"
	"github.com/goverland-labs/inbox-web-api/internal/helpers"
	"github.com/goverland-labs/inbox-web-api/internal/rest/forms/tools"
	"github.com/goverland-labs/inbox-web-api/internal/rest/request"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	address := mux.Vars(r)["address"]
	enslist, err := s.coreclient.GetEnsNames(r.Context(), coresdk.GetEnsNamesRequest{
		Addresses: []string{address},
	})
	if err != nil {
		log.Error().Err(err).Msg("get public profile info")
		response.SendEmpty(w, http.StatusInternalServerError)

		return
	}
	alias := address
	var ensName *string
	if len(enslist.EnsNames) > 0 && enslist.EnsNames[0].Name != "" {
		ensName = helpers.Ptr(enslist.EnsNames[0].Name)
		alias = enslist.EnsNames[0].Name
	}
	user := common.User{
		Address:      common.UserAddress(address),
		ResolvedName: ensName,
		Avatars:      common.GenerateProfileAvatars(alias),
	}

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Str("address", address).
		Msg("route execution")

	response.SendJSON(w, http.StatusOK, &user)
}

func (s *Server) getUserVotes(w http.ResponseWriter, r *http.Request) {
	offset, limit, err := request.ExtractPagination(r)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	var proposalWithVotes []proposal.Proposal
	total := 0
	session, _ := appctx.ExtractUserSession(r.Context())
	address, ok := s.getUserAddress(session)
	if !ok {
		proposalWithVotes = make([]proposal.Proposal, 0)
	} else {
		resp, err := s.coreclient.GetUserVotes(r.Context(), address, coresdk.GetUserVotesRequest{
			Offset: offset,
			Limit:  limit,
		})
		if err != nil {
			log.Error().Err(err).Msgf("get user votes by address: %s", address)

			response.SendEmpty(w, http.StatusInternalServerError)
			return
		}
		proposalWithVotes = make([]proposal.Proposal, 0)
		if len(resp.Items) != 0 {
			userProposals, err := s.collectProposals(resp.Items, r.Context())
			if err != nil {
				response.SendError(w, http.StatusBadRequest, err.Error())
				return
			}
			list := ConvertVoteToInternal(resp.Items)
			for _, info := range list {
				p, ok := userProposals[info.ProposalID]
				if !ok {
					continue
				}
				p.UserVote = helpers.Ptr(info)
				proposalWithVotes = append(proposalWithVotes, p)
			}
		}
		total = resp.TotalCnt
	}
	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(proposalWithVotes)).
		Int("total", total).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, offset, limit, total)
	response.SendJSON(w, http.StatusOK, &proposalWithVotes)
}

func (s *Server) getPublicUserVotes(w http.ResponseWriter, r *http.Request) {
	offset, limit, err := request.ExtractPagination(r)
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	address := mux.Vars(r)["address"]

	var daoID *string
	if daoIDStr := r.URL.Query().Get("dao"); daoIDStr != "" {
		daoID = helpers.Ptr(daoIDStr)
	}

	resp, err := s.coreclient.GetUserVotes(r.Context(), address, coresdk.GetUserVotesRequest{
		Offset: offset,
		Limit:  limit,
		DaoID:  daoID,
	})
	if err != nil {
		log.Error().Err(err).Msgf("get user votes by address: %s", address)

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}
	proposalWithVotes := make([]proposal.Proposal, 0)
	if len(resp.Items) != 0 {
		userProposals, err := s.collectProposals(resp.Items, r.Context())
		if err != nil {
			response.SendError(w, http.StatusBadRequest, err.Error())
			return
		}
		proposalIds := make([]string, 0)
		for _, info := range resp.Items {
			proposalIds = append(proposalIds, info.ProposalID)
		}

		list := ConvertVoteToInternal(resp.Items)
		session, _ := appctx.ExtractUserSession(r.Context())
		meAddress, meAddressExists := s.getUserAddress(session)

		meVotes := make(map[string]proposal.Vote)
		if meAddressExists {
			meResp, _ := s.coreclient.GetUserVotes(r.Context(), meAddress, coresdk.GetUserVotesRequest{
				Limit:       limit,
				ProposalIDs: proposalIds,
			})
			meList := ConvertVoteToInternal(meResp.Items)
			for _, info := range meList {
				meVotes[info.ProposalID] = info
			}
		}
		for _, info := range list {
			p, ok := userProposals[info.ProposalID]
			if !ok {
				continue
			}
			p.PublicUserVote = helpers.Ptr(info)
			v, ok := meVotes[info.ProposalID]
			if ok {
				p.UserVote = helpers.Ptr(v)
			}
			proposalWithVotes = append(proposalWithVotes, p)
		}
	}

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(proposalWithVotes)).
		Int("total", resp.TotalCnt).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, offset, limit, resp.TotalCnt)
	response.SendJSON(w, http.StatusOK, &proposalWithVotes)
}

func (s *Server) getAddressVotingPower(w http.ResponseWriter, r *http.Request) {
	f, verr := tools.NewVotingPowerForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	var vp internaltools.VotingPower
	for _, addr := range f.Addresses {
		score := 0
		// todo maybe create batch method, but seems in our case generally we will have one address to check
		user, err := s.userClient.GetUser(r.Context(), &inboxapi.GetUserRequest{
			Address: addr,
		})
		if err != nil {
			log.Error().Err(err).Msg("get user by address")
		} else if user.GetRole() == inboxapi.UserRole_USER_ROLE_REGULAR {
			score = 1
		}

		vp.Score = append(vp.Score, internaltools.VotingPowerScore{
			Score:   score,
			Address: addr,
		})
	}

	response.SendJSON(w, http.StatusOK, &vp)
}

func (s *Server) getParticipatedDaos(w http.ResponseWriter, r *http.Request) {
	address := mux.Vars(r)["address"]
	list, err := s.coreclient.GetUserParticipatedDaos(r.Context(), address)
	if err != nil {
		log.Error().Err(err).Msg("get participated daos")
		response.SendEmpty(w, http.StatusInternalServerError)

		return
	}
	if list.TotalCnt == 0 {
		response.AddPaginationHeaders(w, r, 0, 0, 0)
		response.SendJSON(w, http.StatusOK, &[]*internaldao.DAO{})

		return
	}
	daoIds := make([]string, 0)
	for _, info := range list.Ids {
		daoIds = append(daoIds, info.String())
	}
	daolist, err := s.daoService.GetDaoList(r.Context(), internaldao.DaoListRequest{
		IDs:   daoIds,
		Limit: len(daoIds),
	})

	daos := helpers.WrapDAOsIpfsLinks(daolist.Items)
	session, _ := appctx.ExtractUserSession(r.Context())
	daos = enrichSubscriptionInfo(session, daos)

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(daos)).
		Msg("route execution")

	response.AddPaginationHeaders(w, r, 0, len(daos), daolist.TotalCnt)
	response.SendJSON(w, http.StatusOK, &daos)
}

func (s *Server) collectProposals(votes []coreproposal.Vote, ctx context.Context) (map[string]proposal.Proposal, error) {
	daoIds := make([]string, 0)
	proposalIds := make([]string, 0)
	for _, info := range votes {
		daoIds = append(daoIds, info.DaoID.String())
		proposalIds = append(proposalIds, info.ProposalID)
	}
	slices.Sort(daoIds)
	daoIds = slices.Compact(daoIds)
	daolist, err := s.daoService.GetDaoList(ctx, internaldao.DaoListRequest{
		IDs:   daoIds,
		Limit: len(daoIds),
	})
	if err != nil {
		log.Error().Err(err).Msg("get dao list by IDs")

		return nil, err
	}
	pc := len(proposalIds)
	proposalItems := make([]coreproposal.Proposal, 0)
	for pc > 0 {
		limit := pc
		ids := proposalIds
		const MAX_PROPOSALS_BY_REQUEST = 80
		if pc > MAX_PROPOSALS_BY_REQUEST {
			limit = MAX_PROPOSALS_BY_REQUEST
			ids = proposalIds[:MAX_PROPOSALS_BY_REQUEST]
			proposalIds = proposalIds[MAX_PROPOSALS_BY_REQUEST:]
		}
		proposallist, err := s.coreclient.GetProposalList(ctx, coresdk.GetProposalListRequest{
			ProposalIDs: ids,
			Limit:       limit})
		if err != nil {
			log.Error().Err(err).Msg("get proposal list")

			return nil, err
		}
		proposalItems = append(proposalItems, proposallist.Items...)
		pc = pc - MAX_PROPOSALS_BY_REQUEST
	}

	daos := make(map[string]*internaldao.DAO)
	for _, info := range daolist.Items {
		daos[info.ID.String()] = info
	}

	proposals := make([]proposal.Proposal, len(proposalItems))
	for i, info := range proposalItems {
		di, ok := daos[info.DaoID.String()]
		if !ok {
			log.Error().Msg("dao not found")

			return nil, err
		}
		proposals[i] = convertProposalToInternal(&info, di)
	}
	session, _ := appctx.ExtractUserSession(ctx)
	proposals = enrichProposalsSubscriptionInfo(session, proposals)
	proposals = helpers.WrapProposalsIpfsLinks(proposals)

	userProposals := make(map[string]proposal.Proposal)
	for _, info := range proposals {
		userProposals[info.ID] = info
	}
	return userProposals, nil
}

func (s *Server) getMeCanVote(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	proposals, err := s.userClient.GetUserCanVoteProposals(r.Context(), &inboxapi.GetUserCanVoteProposalsRequest{
		UserId: session.UserID.String(),
	})
	if err != nil {
		log.Error().Err(err).Msg("get user can vote proposals")

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	if len(proposals.GetProposalIds()) == 0 {
		log.Info().
			Str("user_id", session.UserID.String()).
			Int("count", 0).
			Msg("me can vote")

		response.SendJSON(w, http.StatusOK, helpers.Ptr([]proposal.Proposal{}))
		return
	}

	proposalList, err := s.coreclient.GetProposalList(r.Context(), coresdk.GetProposalListRequest{
		ProposalIDs: proposals.GetProposalIds(),
		Limit:       len(proposals.GetProposalIds()),
	})
	if err != nil {
		log.Error().Err(err).Msg("get proposal list")

		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// todo: use caching for getting dao
	daoIds := make([]string, 0)
	for _, info := range proposalList.Items {
		daoIds = append(daoIds, info.DaoID.String())
	}
	daolist, err := s.daoService.GetDaoList(r.Context(), internaldao.DaoListRequest{
		IDs:   daoIds,
		Limit: len(daoIds),
	})
	if err != nil {
		log.Error().Err(err).Msg("get dao list by IDs")

		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	daos := make(map[string]*internaldao.DAO)
	for _, info := range daolist.Items {
		daos[info.ID.String()] = info
	}

	list := make([]proposal.Proposal, len(proposalList.Items))
	for i, info := range proposalList.Items {
		di, ok := daos[info.DaoID.String()]
		if !ok {
			log.Error().Msg("dao not found")

			response.SendError(w, http.StatusBadRequest, fmt.Sprintf("dao not found: %s", info.DaoID))
			return
		}
		list[i] = convertProposalToInternal(&info, di)
	}

	list = s.enrichProposalsVotesInfo(r.Context(), session, list)
	list = helpers.WrapProposalsIpfsLinks(list)

	proposalsWithoutVotes := make([]proposal.Proposal, 0, len(list))
	for _, p := range list {
		if p.UserVote != nil {
			continue
		}

		if !p.IsActive() {
			continue
		}

		proposalsWithoutVotes = append(proposalsWithoutVotes, p)
	}

	log.Info().
		Str("user_id", session.UserID.String()).
		Int("count_filtered", len(proposalsWithoutVotes)).
		Int("total", proposalList.TotalCnt).
		Msg("me can vote")

	response.SendJSON(w, http.StatusOK, &proposalsWithoutVotes)
}

func (s *Server) getVoteNow(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	var featured bool
	if featuredStr := r.URL.Query().Get("featured"); featuredStr != "" {
		featured = featuredStr == "true"
	}

	subscribtions, err := s.subclient.ListSubscriptions(r.Context(), &inboxapi.ListSubscriptionRequest{
		SubscriberId: session.UserID.String(),
	})
	if err != nil {
		log.Error().Err(err).Msg("get user subscriptions")

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	if len(subscribtions.Items) == 0 {
		log.Info().
			Str("user_id", session.UserID.String()).
			Int("count", 0).
			Msg("vote now")

		response.SendJSON(w, http.StatusOK, helpers.Ptr([]proposal.Proposal{}))
		return
	}

	daoIdsForReq := make([]string, 0, len(subscribtions.GetItems()))
	for _, info := range subscribtions.GetItems() {
		daoIdsForReq = append(daoIdsForReq, info.DaoId)
	}

	proposalList, err := s.coreclient.GetProposalList(r.Context(), coresdk.GetProposalListRequest{
		Dao:        strings.Join(daoIdsForReq, ","),
		OnlyActive: true,
		Offset:     0,
		Limit:      1000,
	})
	if err != nil {
		log.Error().Err(err).Msg("get proposal list")

		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// todo: use caching for getting dao
	daoIds := make([]string, 0)
	for _, info := range proposalList.Items {
		daoIds = append(daoIds, info.DaoID.String())
	}
	daolist, err := s.daoService.GetDaoList(r.Context(), internaldao.DaoListRequest{
		IDs:   daoIds,
		Limit: len(daoIds),
	})
	if err != nil {
		log.Error().Err(err).Msg("get dao list by IDs")

		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	daos := make(map[string]*internaldao.DAO)
	for _, info := range daolist.Items {
		daos[info.ID.String()] = info
	}

	list := make([]proposal.Proposal, len(proposalList.Items))
	for i, info := range proposalList.Items {
		di, ok := daos[info.DaoID.String()]
		if !ok {
			log.Error().Msg("dao not found")

			response.SendError(w, http.StatusBadRequest, fmt.Sprintf("dao not found: %s", info.DaoID))
			return
		}
		list[i] = convertProposalToInternal(&info, di)
	}

	list = s.enrichProposalsVotesInfo(r.Context(), session, list)
	list = helpers.WrapProposalsIpfsLinks(list)

	resultProposals := make([]proposal.Proposal, 0, len(list))
	for _, p := range list {
		if p.UserVote != nil {
			continue
		}

		if !p.IsActive() {
			continue
		}

		resultProposals = append(resultProposals, p)
	}

	total := proposalList.TotalCnt
	if featured {
		const maxFeaturedProposals = 3
		usedDaos := make(map[uuid.UUID]struct{})

		featuredProposals := make([]proposal.Proposal, 0, maxFeaturedProposals)
		for _, p := range resultProposals {
			if len(featuredProposals) >= maxFeaturedProposals {
				break
			}

			if _, ok := usedDaos[p.DAO.ID]; ok {
				continue
			}

			featuredProposals = append(featuredProposals, p)
			usedDaos[p.DAO.ID] = struct{}{}
		}

		resultProposals = featuredProposals
		total = len(resultProposals)
	}

	log.Info().
		Str("user_id", session.UserID.String()).
		Int("count_filtered", len(resultProposals)).
		Int("total", total).
		Msg("vote now")

	response.SendJSON(w, http.StatusOK, &resultProposals)
}

func (s *Server) getUserAddress(session auth.Session) (address string, exist bool) {
	if session == auth.EmptySession {
		return "", false
	}
	profileInfo, err := s.authService.GetProfileInfo(session.UserID)
	if err != nil || profileInfo.Account == nil {
		return "", false
	}

	ad := profileInfo.Account.Address
	if ad == "" {
		return "", false
	}
	return ad, true
}

func (s *Server) getRecommendedDao(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	available, err := s.userClient.GetAvailableDaoByWallet(r.Context(), &inboxapi.GetAvailableDaoByWalletRequest{
		UserId: session.UserID.String(),
	})
	if err != nil {
		switch status.Code(err) {
		case codes.InvalidArgument, codes.FailedPrecondition:
			response.SendError(w, http.StatusBadRequest, err.Error())
		default:
			w.WriteHeader(http.StatusInternalServerError)
			response.SendError(w, http.StatusInternalServerError, err.Error())
		}

		return
	}

	daoIDs := make([]string, 0, len(available.DaoUuids))
	for i := range available.DaoUuids {
		daoIDs = append(daoIDs, available.DaoUuids[i])
	}

	daoList, err := s.daoService.GetDaoByIDs(r.Context(), daoIDs...)
	if err != nil {
		log.Error().Err(err).Msg("get dao list by IDs")

		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	list := make([]*internaldao.DAO, 0, len(daoList))
	for _, di := range daoList {
		list = append(list, di)
	}

	// sort by popularity index desc
	sort.Slice(list, func(i, j int) bool {
		return list[i].PopularityIndex > list[j].PopularityIndex
	})

	list = helpers.WrapDAOsIpfsLinks(list)
	list = enrichSubscriptionInfo(session, list)

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Int("count", len(list)).
		Msg("route execution")

	response.SendJSON(w, http.StatusOK, &list)
}
