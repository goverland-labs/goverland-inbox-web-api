package ipfs

import (
	"fmt"
	"regexp"
	"strings"
)

const ipfsLinkTemplate = "https://ipfs.io/ipfs/%s"
const ipfsResolveURL = "https://snapshot.4everland.link/ipfs/"
const ipfsDAOImageLinkTemplate = "https://cdn.stamp.fyi/space/%s?s=180"

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

func ReplaceLinksInText(text string) string {
	return strings.ReplaceAll(text, "ipfs://", ipfsResolveURL)
}
