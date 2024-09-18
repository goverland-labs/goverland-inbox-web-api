package dao

import (
	"strings"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestAddressConvert(t *testing.T) {
	bytes := ethcommon.Hex2Bytes(strings.TrimPrefix("0x91e2E2D26076C8A1EaDb69273605c16ef01928ce", "0x"))
	converted := ethcommon.LeftPadBytes(bytes, 32)
	assert.Equal(t, "00000000000000000000000091e2e2d26076c8a1eadb69273605c16ef01928ce", ethcommon.Bytes2Hex(converted))
}
