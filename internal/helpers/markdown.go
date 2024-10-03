package helpers

import (
	"regexp"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var renderer = html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags | html.HrefTargetBlank})

var inlineImage = regexp.MustCompile(`\n(!\[.*\]\(.*\))\n`)

func CompileMarkdown(md string) string {
	tree := markdown.Parse([]byte(md), parser.New())

	return string(markdown.Render(tree, renderer))
}

func ReplaceInlineImages(text string) string {
	return inlineImage.ReplaceAllString(text, "\n\n$1\n\n")
}
