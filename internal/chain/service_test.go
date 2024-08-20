package chain

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestGetChainsInfo(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Error(err)
	}

	result, err := service.GetGasPriceHex(EthChainID)
	if err != nil {
		t.Error(err)
	}

	spew.Dump(result)
}
