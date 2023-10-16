package ipfs

import (
	"fmt"
	"regexp"
)

const ipfsLinkTemplate = "https://ipfs.io/ipfs/%s"
const ipfsDAOImageLinkTemplate = "https://cdn.stamp.fyi/space/%s?s=90"

var linkRE = regexp.MustCompile(`^ipfs://([a-zA-Z0-9]+)$`)

func WrapLink(link string) string {
	matches := linkRE.FindStringSubmatch(link)
	if len(matches) != 2 {
		return link
	}

	return fmt.Sprintf(ipfsLinkTemplate, matches[1])
}

func WrapDAOImageLink(ensName string) string {
	return fmt.Sprintf(ipfsDAOImageLinkTemplate, ensName)
}
