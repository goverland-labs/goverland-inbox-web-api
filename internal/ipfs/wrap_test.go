package ipfs

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestWrapLink(t *testing.T) {
	convey.Convey("wrapping ipfs links", t, func() {
		wrapped := WrapLink("ipfs://bafkreihd7yqih3qfzxnj5cibp3yyqb5tjkamgqyz3kn3yiraumcyz4y5au")
		convey.So(wrapped, convey.ShouldEqual, "https://ipfs.io/ipfs/bafkreihd7yqih3qfzxnj5cibp3yyqb5tjkamgqyz3kn3yiraumcyz4y5au")
	})
}
