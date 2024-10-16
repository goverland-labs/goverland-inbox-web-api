package rest

import (
	"github.com/goverland-labs/inbox-web-api/internal/chain"
	"github.com/goverland-labs/inbox-web-api/internal/rest/response"
)

var chainResponseErrors = map[error]func(err error) response.Error{
	chain.ErrChainRequestUnreachable: func(err error) response.Error {
		return response.NewUnprocessableError(err, "Sorry, chain is busy. Please try again in a moment.")
	},
	chain.ErrEstimateFee: func(err error) response.Error {
		return response.NewUnprocessableError(err, "Sorry, we couldn't estimate the transaction, is can be duplicate delegation or something else.")
	},
}
