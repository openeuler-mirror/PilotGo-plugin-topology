package traversal

// GremlinTraversalExtension interface for Gremlin language extension
type GremlinTraversalExtension interface {
	ScanIdent(s string) (Token, bool)
	ParseStep(t Token, p GremlinTraversalContext) (GremlinTraversalStep, error)
}
