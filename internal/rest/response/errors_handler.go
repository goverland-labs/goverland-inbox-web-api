package response

import (
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/helpers"
)

func HandleError(err Error, w http.ResponseWriter) {
	if err == nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Msg("unhandled error")

		return
	}

	SendJSON(w, err.GetHTTPStatus(), helpers.Ptr(ParseError(err)))
}

// ParseError determines the error type and creates a map with the error description.
func ParseError(err Error) map[string]interface{} {
	if err == nil {
		return nil
	}

	switch e := err.(type) { // nolint:gocritic
	case *ValidationError:
		return map[string]interface{}{
			"errors":  e.Errors(),
			"message": e.PublicMessage(),
		}

	case *UnprocessableEntityValidationError:
		return map[string]interface{}{
			"errors":  e.Errors(),
			"message": e.PublicMessage(),
		}

	case *RateLimitedError:
		return map[string]interface{}{
			"message":     e.PublicMessage(),
			"retry_after": e.RetryAfter,
		}

	case *PermissionDeniedError:
		return map[string]interface{}{
			"errors":  e.Errors(),
			"message": e.PublicMessage(),
		}

	default:
		return map[string]interface{}{
			"message": err.PublicMessage(),
		}
	}
}
