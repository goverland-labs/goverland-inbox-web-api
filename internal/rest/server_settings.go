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
		UserId: session.UserID.String(),
		Token:  f.Token,
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
		UserId: session.UserID.String(),
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
		UserId: session.UserID.String(),
	})
	if err != nil {
		log.Error().Err(err).Msgf("remove push token for user: %s", session.UserID.String())

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	response.SendEmpty(w, http.StatusOK)
}
