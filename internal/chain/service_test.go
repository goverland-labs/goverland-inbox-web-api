package chain

import (
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/goverland-labs/inbox-web-api/internal/config"
)

func TestGetChainsInfo(t *testing.T) {
	service, err := NewService(config.Chain{})
	if err != nil {
		t.Error(err)
	}

	result, err := service.GetGasPriceHex(1)
	if err != nil {
		t.Error(err)
	}

	spew.Dump(result)
}
