package rest

import (
	"github.com/goverland-labs/inbox-web-api/internal/chain"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

var chainResponseErrors = map[error]func(err error) response.Error{
	chain.ErrChainRequestUnreachable: func(err error) response.Error {
		return response.NewUnprocessableError(err, "Sorry, we couldn't request delegation data. Please try again later.")
	},
}
