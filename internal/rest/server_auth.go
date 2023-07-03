package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/inbox-web-api/internal/rest/forms/auth"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

func (s *Server) authByDevice(w http.ResponseWriter, r *http.Request) {
	f, verr := auth.NewByDeviceForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	session, err := s.authStorage.Guest(f.DeviceID)
	if err != nil {
		log.Error().Err(err).Msg("guest session")

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	s.getSubscriptions(session.ID)

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Str("session_id", session.ID.String()).
		Msg("route execution")

	response.SendJSON(w, http.StatusOK, &session)
}
