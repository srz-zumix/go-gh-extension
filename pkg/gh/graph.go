package gh

// Edge represents a directed connection between two GitHub entities, such as repositories, teams, or users.
type GraphEdge struct {
	From any
	To   any
}

// GetFromName returns the name of the source entity in the edge
func (e GraphEdge) GetFromName() string {
	return GetObjectName(e.From)
}

// GetToName returns the name of the target entity in the edge
func (e GraphEdge) GetToName() string {
	return GetObjectName(e.To)
}
