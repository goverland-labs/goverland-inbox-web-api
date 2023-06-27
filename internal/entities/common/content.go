package common

const (
	Markdown ContentType = "markdown"
	HTML     ContentType = "html"
)

type ContentType string

type Content struct {
	Type ContentType `json:"type"`
	Body string      `json:"body"`
}
