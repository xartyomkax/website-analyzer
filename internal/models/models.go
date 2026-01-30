package models

// LinkType represents the category of a link
type LinkType int

const (
	LinkTypeInternal LinkType = iota
	LinkTypeExternal
	LinkTypeInvalid
)

func (lt LinkType) String() string {
	switch lt {
	case LinkTypeInternal:
		return "internal"
	case LinkTypeExternal:
		return "external"
	default:
		return "invalid"
	}
}

// Link represents a hyperlink found in the document
type Link struct {
	URL  string   `json:"url"`
	Type LinkType `json:"type"`
}

// AnalysisResult contains all analysis data for a webpage
type AnalysisResult struct {
	URL               string         `json:"url"`
	HTMLVersion       string         `json:"html_version"`
	Title             string         `json:"title"`
	Headings          map[string]int `json:"headings"`
	InternalLinks     int            `json:"internal_links"`
	ExternalLinks     int            `json:"external_links"`
	InaccessibleLinks []LinkError    `json:"inaccessible_links"`
	HasLoginForm      bool           `json:"has_login_form"`
}

// LinkError represents a link that could not be accessed
type LinkError struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error"`
}
