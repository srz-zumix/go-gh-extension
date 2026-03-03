package gh

// Edge represents a directed connection between two GitHub entities, such as repositories, teams, or users.
type GraphEdge struct {
	From any
	To   any
	Host string // GitHub host for URL generation (e.g. "github.com" or GHES hostname)
}

// GetFromName returns the name of the source entity in the edge
func (e GraphEdge) GetFromName() string {
	return GetObjectName(e.From)
}

// GetToName returns the name of the target entity in the edge
func (e GraphEdge) GetToName() string {
	return GetObjectName(e.To)
}

// GetFromURL returns a browsable URL for the source entity in the edge
func (e GraphEdge) GetFromURL() string {
	return GetObjectHTMLURL(e.From, e.Host)
}

// GetToURL returns a browsable URL for the target entity in the edge
func (e GraphEdge) GetToURL() string {
	return GetObjectHTMLURL(e.To, e.Host)
}
