package rest

import (
	"net/http"

	"github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/entities/settings"
	settingsform "github.com/goverland-labs/inbox-web-api/internal/rest/forms/settings"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
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

	f, verr := settingsform.NewStoreSettingsForm().ParseAndValidate(r)
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

func fillStoreDAOSettingsRequest(in *settingsform.StoreSettingsForm) []inboxapi.PUSH_SETTINGS_DAO {
	var result []inboxapi.PUSH_SETTINGS_DAO
	if in.Dao.NewProposalCreated {
		result = append(result, inboxapi.PUSH_SETTINGS_DAO_PUSH_SETTINGS_NEW_PROPOSAL_CREATED)
	}

	if in.Dao.QuorumReached {
		result = append(result, inboxapi.PUSH_SETTINGS_DAO_PUSH_SETTINGS_QUORUM_REACHED)
	}

	if in.Dao.VoteFinished {
		result = append(result, inboxapi.PUSH_SETTINGS_DAO_PUSH_SETTINGS_VOTE_FINISHED)
	}

	if in.Dao.VoteFinishesSoon {
		result = append(result, inboxapi.PUSH_SETTINGS_DAO_PUSH_SETTINGS_VOTE_FINISHES_SOON)
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
	if err != nil {
		log.Error().Err(err).Msgf("remove push token for user: %s", session.UserID.String())

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	response.SendJSON(w, http.StatusOK, convertPushDetailsToInternal(details))
}

func convertPushDetailsToInternal(in *inboxapi.GetPushDetailsResponse) *settings.Details {
	details := &settings.Details{}

	for _, option := range in.GetDao() {
		switch option {
		case inboxapi.PUSH_SETTINGS_DAO_PUSH_SETTINGS_VOTE_FINISHES_SOON:
			details.Dao.VoteFinishesSoon = true
		case inboxapi.PUSH_SETTINGS_DAO_PUSH_SETTINGS_VOTE_FINISHED:
			details.Dao.VoteFinished = true
		case inboxapi.PUSH_SETTINGS_DAO_PUSH_SETTINGS_QUORUM_REACHED:
			details.Dao.QuorumReached = true
		case inboxapi.PUSH_SETTINGS_DAO_PUSH_SETTINGS_NEW_PROPOSAL_CREATED:
			details.Dao.NewProposalCreated = true
		}
	}

	return details
}
