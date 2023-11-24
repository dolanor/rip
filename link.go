package rip

// Link represents a RFC8288 web link.
type Link struct {
	// HRef is a URI-reference [RFC3986 Section 4.1] pointing to the link’s target.
	HRef string `json:"href,omitempty"`

	// Rel indicates the link’s relation type. The string MUST be a valid link relation type.
	Rel string `json:"rel,omitempty"`

	// DescribedBy is a link to a description document (e.g. OpenAPI or JSON Schema) for the link target.
	DescribedBy *Link `json:"describedby,omitempty"`

	// Title serves as a label for the destination of a link such that it can be used as a human-readable identifier (e.g., a menu entry).
	Title string `json:"title,omitempty"`

	// Type indicates the media type of the link’s target.
	Type string `json:"type,omitempty"`

	// HRefLang indicates the language(s) of the link’s target. An array of strings indicates that the link’s target is available in multiple languages. Each string MUST be a valid language tag [RFC5646].
	HRefLang []string `json:"hreflang,omitempty"`
}
