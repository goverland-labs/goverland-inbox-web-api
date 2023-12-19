package rest

import (
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/spruceid/siwe-go"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	authsrv "github.com/goverland-labs/inbox-web-api/internal/auth"
	"github.com/goverland-labs/inbox-web-api/internal/rest/forms/auth"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

func (s *Server) guestAuth(w http.ResponseWriter, r *http.Request) {
	f, verr := auth.NewGuestAuthForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	guestInfo, err := s.authService.Guest(r.Context(), authsrv.GuestSessionRequest{
		SessionRequest: authsrv.SessionRequest{
			DeviceID:   f.DeviceID,
			DeviceName: f.DeviceName,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("guest session")

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	s.getSubscriptions(guestInfo.Session.ID)

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Str("session_id", guestInfo.Session.ID.String()).
		Msg("route execution")

	response.SendJSON(w, http.StatusOK, &guestInfo.AuthInfo)
}

func (s *Server) siweAuth(w http.ResponseWriter, r *http.Request) {
	f, verr := auth.NewSiweAuthForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	siweMessage, err := siwe.ParseMessage(f.Message)
	if err != nil {
		log.Error().Err(err).Msg("siwe parse message")

		response.SendError(w, http.StatusBadRequest, "invalid siwe message")
		return
	}

	pubKey, err := siweMessage.Verify(f.Signature, nil, nil, nil)
	if err != nil {
		log.Error().Err(err).Msg("siwe verify")

		response.SendError(w, http.StatusBadRequest, "invalid siwe signature or expired")
		return
	}
	encodedAddress := crypto.PubkeyToAddress(*pubKey)

	if encodedAddress != f.Address {
		log.Error().Err(err).Msg("address is not related to sign")

		response.SendError(w, http.StatusBadRequest, "address is not related to sign")
		return
	}

	regularInfo, err := s.authService.Regular(r.Context(), authsrv.RegularSessionRequest{
		SessionRequest: authsrv.SessionRequest{
			DeviceID:   f.DeviceID,
			DeviceName: f.DeviceName,
		},
		Address: f.Address.Hex(),
	})
	if err != nil {
		log.Error().Err(err).Msg("guest session")

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	s.getSubscriptions(regularInfo.Session.ID)

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Str("session_id", regularInfo.Session.ID.String()).
		Msg("route execution")

	response.SendJSON(w, http.StatusOK, &regularInfo.AuthInfo)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	err := s.authService.Logout(session.ID)
	if err != nil {
		log.Error().Err(err).Msg("logout")
		response.SendEmpty(w, http.StatusInternalServerError)

		return
	}

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Str("session_id", session.ID.String()).
		Msg("route execution")

	response.SendEmpty(w, http.StatusNoContent)
}

func (s *Server) getMe(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	profileInfo, err := s.authService.GetProfileInfo(session.UserID)
	if err != nil {
		log.Error().Err(err).Msg("get profile info")
		response.SendEmpty(w, http.StatusInternalServerError)

		return
	}

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Str("session_id", session.ID.String()).
		Msg("route execution")

	response.SendJSON(w, http.StatusOK, &profileInfo)
}

func (s *Server) deleteMe(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	err := s.authService.DeleteUser(session.UserID)
	if err != nil {
		log.Error().Err(err).Msg("delete me")
		response.SendEmpty(w, http.StatusInternalServerError)

		return
	}

	log.Info().
		Str("route", mux.CurrentRoute(r).GetName()).
		Str("session_id", session.ID.String()).
		Msg("route execution")

	response.SendEmpty(w, http.StatusNoContent)
}
