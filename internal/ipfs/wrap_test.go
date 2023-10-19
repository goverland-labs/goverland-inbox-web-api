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

func TestReplaceLinksInText(t *testing.T) {
	convey.Convey("replacing ipfs links in text", t, func() {
		convey.Convey("text without links should not be changed", func() {
			text := `
				Issue Statement**
				Stargate Finance has grown exponentially in 2023 hitting major protocol milestones, 
				including over $3.4B in monthly transaction volume and over $2M in monthly protocol fees generated.
			`
			convey.So(ReplaceLinksInText(text), convey.ShouldEqual, text)
		})

		convey.Convey("text with one link should replace that link", func() {
			text := `
				The Stargate Foundation will ![image](ipfs://bafkreih657l4daehca3f56mya3b7juvmrwsokblacjmpfk6tllqupdw3rm)
				**Summary**
				The Stargate Foundation does important work to support the Protocol and the mandate of the Stargate DAO
			`

			expected := `
				The Stargate Foundation will ![image](https://gateway.4everland.link/ipfs/bafkreih657l4daehca3f56mya3b7juvmrwsokblacjmpfk6tllqupdw3rm)
				**Summary**
				The Stargate Foundation does important work to support the Protocol and the mandate of the Stargate DAO
			`

			convey.So(ReplaceLinksInText(text), convey.ShouldEqual, expected)
		})

		convey.Convey("text with links should replace all links", func() {
			text := `
				The Stargate Foundation will ![image](ipfs://bafkreih657l4daehca3f56mya3b7juvmrwsokblacjmpfk6tllqupdw3rm)
				**Summary** ipfs://bafkreih657l4daehca3f56mya3b7juvmrwsokblacjmpfk6tllqupdw3rm
				The Stargate Foundation does important work to support the Protocol and the mandate of the Stargate DAO
				ipfs://bafkreih657l4daehca3f56mya3b7juvmrwsokblacjmpfk6tllqupdw3rm
			`

			expected := `
				The Stargate Foundation will ![image](https://gateway.4everland.link/ipfs/bafkreih657l4daehca3f56mya3b7juvmrwsokblacjmpfk6tllqupdw3rm)
				**Summary** https://gateway.4everland.link/ipfs/bafkreih657l4daehca3f56mya3b7juvmrwsokblacjmpfk6tllqupdw3rm
				The Stargate Foundation does important work to support the Protocol and the mandate of the Stargate DAO
				https://gateway.4everland.link/ipfs/bafkreih657l4daehca3f56mya3b7juvmrwsokblacjmpfk6tllqupdw3rm
			`

			convey.So(ReplaceLinksInText(text), convey.ShouldEqual, expected)
		})

		convey.Convey("twice replace should be idempotent", func() {
			text := `
				The Stargate Foundation will ![image](ipfs://bafkreih657l4daehca3f56mya3b7juvmrwsokblacjmpfk6tllqupdw3rm)
				**Summary**
				The Stargate Foundation does important work to support the Protocol and the mandate of the Stargate DAO
			`

			expected := `
				The Stargate Foundation will ![image](https://gateway.4everland.link/ipfs/bafkreih657l4daehca3f56mya3b7juvmrwsokblacjmpfk6tllqupdw3rm)
				**Summary**
				The Stargate Foundation does important work to support the Protocol and the mandate of the Stargate DAO
			`

			convey.So(ReplaceLinksInText(text), convey.ShouldEqual, expected)
			convey.So(ReplaceLinksInText(expected), convey.ShouldEqual, expected)
		})
	})
}
