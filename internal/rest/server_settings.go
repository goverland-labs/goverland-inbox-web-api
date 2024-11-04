package rest

import (
	"net/http"

	"github.com/goverland-labs/goverland-inbox-api-protocol/protobuf/inboxapi"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/appctx"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/entities/settings"
	settingsform "github.com/goverland-labs/goverland-inbox-web-api/internal/rest/forms/settings"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest/response"
)

func (s *Server) storePushToken(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := settingsform.NewStoreTokenForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	_, err := s.settings.AddPushToken(r.Context(), &inboxapi.AddPushTokenRequest{
		UserId:     session.UserID.String(),
		Token:      f.Token,
		DeviceUuid: session.DeviceUUID,
	})
	if err != nil {
		log.Error().Err(err).Msgf("store token for user: %s", session.UserID.String())

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	response.SendEmpty(w, http.StatusOK)
}

func (s *Server) tokenExists(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	resp, err := s.settings.PushTokenExists(r.Context(), &inboxapi.PushTokenExistsRequest{
		UserId:     session.UserID.String(),
		DeviceUuid: session.DeviceUUID,
	})
	if err != nil {
		log.Error().Err(err).Msgf("check token exists for user: %s", session.UserID.String())

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	response.SendJSON(w, http.StatusOK, &settings.PushTokenExists{Enabled: resp.Exists})
}

func (s *Server) removePushToken(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, err := s.settings.RemovePushToken(r.Context(), &inboxapi.RemovePushTokenRequest{
		UserId:     session.UserID.String(),
		DeviceUuid: session.DeviceUUID,
	})
	if err != nil {
		log.Error().Err(err).Msgf("remove push token for user: %s", session.UserID.String())

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	response.SendEmpty(w, http.StatusOK)
}

func (s *Server) storeSettings(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := settingsform.NewStorePushSettingsForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	_, err := s.settings.SetPushDetails(r.Context(), &inboxapi.SetPushDetailsRequest{
		UserId: session.UserID.String(),
		Dao:    fillStoreDAOSettingsRequest(f),
	})
	if err != nil {
		log.Error().Err(err).Msgf("set push details for user: %s", session.UserID.String())

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	response.SendEmpty(w, http.StatusOK)
}

func fillStoreDAOSettingsRequest(in *settingsform.StorePushSettingsForm) *inboxapi.PushSettingsDao {
	result := &inboxapi.PushSettingsDao{
		NewProposalCreated: in.Dao.NewProposalCreated,
		QuorumReached:      in.Dao.QuorumReached,
		VoteFinishesSoon:   in.Dao.VoteFinishesSoon,
		VoteFinished:       in.Dao.VoteFinished,
	}

	return result
}

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	details, err := s.settings.GetPushDetails(r.Context(), &inboxapi.GetPushDetailsRequest{
		UserId: session.UserID.String(),
	})
	if status.Code(err) == codes.NotFound {
		response.SendEmpty(w, http.StatusNotFound)
		return
	}

	if err != nil {
		log.Error().Err(err).Msgf("remove push token for user: %s", session.UserID.String())

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	response.SendJSON(w, http.StatusOK, convertPushDetailsToInternal(details))
}

func convertPushDetailsToInternal(in *inboxapi.GetPushDetailsResponse) *settings.Details {
	details := &settings.Details{}

	details.Dao.VoteFinishesSoon = in.GetDao().GetVoteFinishesSoon()
	details.Dao.VoteFinished = in.GetDao().GetVoteFinished()
	details.Dao.QuorumReached = in.GetDao().GetQuorumReached()
	details.Dao.NewProposalCreated = in.GetDao().GetNewProposalCreated()

	return details
}

func (s *Server) storeFeedSettings(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := settingsform.NewStoreFeedSettingsForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	_, err := s.settings.SetFeedSettings(r.Context(), &inboxapi.SetFeedSettingsRequest{
		UserId: session.UserID.String(),
		FeedSettings: &inboxapi.FeedSettings{
			ArchiveProposalAfterVote: f.ArchiveProposalAfterVote,
			AutoarchiveAfterDuration: f.AutoarchiveAfterDuration,
		},
	})
	if err != nil {
		log.Error().Err(err).Msgf("set feed settings for user: %s", session.UserID.String())

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	response.SendEmpty(w, http.StatusOK)
}

func (s *Server) getFeedSettings(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	details, err := s.settings.GetFeedSettings(r.Context(), &inboxapi.GetFeedSettingsRequest{
		UserId: session.UserID.String(),
	})
	if status.Code(err) == codes.NotFound {
		response.SendEmpty(w, http.StatusNotFound)
		return
	}

	if err != nil {
		log.Error().Err(err).Msgf("remove push token for user: %s", session.UserID.String())

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	response.SendJSON(w, http.StatusOK, convertFeedDetailsToInternal(details))
}

func convertFeedDetailsToInternal(in *inboxapi.GetFeedSettingsResponse) *settings.FeedSettings {
	details := &settings.FeedSettings{}

	details.AutoarchiveAfterDuration = in.GetFeedSettings().GetAutoarchiveAfterDuration()
	details.ArchiveProposalAfterVote = in.GetFeedSettings().GetArchiveProposalAfterVote()

	return details
}
