package rest

import (
	"net/http"

	"github.com/google/uuid"
	events "github.com/goverland-labs/platform-events/events/inbox"
	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/inbox-web-api/internal/appctx"
	"github.com/goverland-labs/inbox-web-api/internal/rest/forms/pushes"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

func (s *Server) sendCustomPush(w http.ResponseWriter, r *http.Request) {
	session, ok := appctx.ExtractUserSession(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := pushes.NewCustomPushForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	if err := s.publisher.PublishJSON(r.Context(), events.SubjectPushCreated, events.PushPayload{
		Title:         f.Title,
		Body:          f.Body,
		UserID:        uuid.UUID(session.UserID),
		ImageURL:      f.ImageURL,
		CustomPayload: f.CustomPayload,
		Version:       events.PushVersionV2,
	}); err != nil {
		log.Error().
			Err(err).
			Msg("publish push message")

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	log.Info().Msg("push sent to the queue")

	response.SendEmpty(w, http.StatusOK)
}

func (s *Server) markAsClicked(w http.ResponseWriter, r *http.Request) {
	if _, ok := appctx.ExtractUserSession(r.Context()); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f, verr := pushes.NewOnClickForm().ParseAndValidate(r)
	if verr != nil {
		response.HandleError(verr, w)
		return
	}

	if err := s.publisher.PublishJSON(r.Context(), events.SubjectPushClicked, events.PushClickPayload{ID: f.ID}); err != nil {
		log.Error().
			Err(err).
			Msg("mark as clicked push message")

		response.SendEmpty(w, http.StatusInternalServerError)
		return
	}

	response.SendEmpty(w, http.StatusOK)
}
