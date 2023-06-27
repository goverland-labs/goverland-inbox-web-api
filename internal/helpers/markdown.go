package helpers

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var renderer = html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags | html.HrefTargetBlank})

func CompileMarkdown(md string) string {
	tree := markdown.Parse([]byte(md), parser.New())

	return string(markdown.Render(tree, renderer))
}
