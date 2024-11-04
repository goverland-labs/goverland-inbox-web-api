package rest

import (
	"net/http"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/entities/stats"
	"github.com/goverland-labs/goverland-inbox-web-api/internal/rest/response"
)

func (s *Server) getStatsTotals(w http.ResponseWriter, r *http.Request) {
	resp, err := s.coreclient.GetStatsTotals(r.Context())
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, "internal error")

		return
	}

	response.SendJSON(w, http.StatusOK, &stats.Totals{
		Dao: stats.Dao{
			Total:         resp.Dao.Total,
			TotalVerified: resp.Dao.TotalVerified,
		},
		Proposals: stats.Proposals{
			Total: resp.Proposals.Total,
		},
	})
}
