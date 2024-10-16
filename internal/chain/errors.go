package chain

import (
	"errors"
)

var ErrChainRequestUnreachable = errors.New("chain request unreachable")
var ErrEstimateFee = errors.New("cannot estimate fee")
