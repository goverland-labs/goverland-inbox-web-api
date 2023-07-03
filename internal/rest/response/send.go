package response

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

func SendJSON[T any](w http.ResponseWriter, status int, data *T) {
	w.Header().Set("Content-Type", "application/json")

	content, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("unable to marshal response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	_, err = w.Write(content)
	if err != nil {
		log.Error().Err(err).Msg("unable to write response to the client")
	}
}

func SendError(w http.ResponseWriter, status int, message string) {
	SendJSON(w, status, &map[string]interface{}{
		"error": message,
	})
}

func SendEmpty(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
}
