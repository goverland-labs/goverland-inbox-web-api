package appversions

type Info struct {
	Version     string `json:"version"`
	Platform    string `json:"platform"`
	Description string `json:"markdown_description"`
}
