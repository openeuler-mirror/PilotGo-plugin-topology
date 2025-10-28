//go:generate stringer -type=Token $GOFILE

package traversal

// Token represents a lexical token.
type Token int

// Default language token, extension token start at 1000
const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	WS

	// Literals
	IDENT

	// Misc characters
	COMMA
	DOT
	LEFTPARENTHESIS
	RIGHTPARENTHESIS
	STRING
	NUMBER

	// Keywords
	G
	V
	E
	HAS
	HASKEY
	HASNOT
	HASEITHER
	OUT
	IN
	OUTV
	INV
	BOTHV
	OUTE
	INE
	BOTHE
	DEDUP
	WITHIN
	WITHOUT
	METADATA
	SHORTESTPATHTO
	NE
	NEE
	BOTH
	CONTEXT
	REGEX
	LT
	GT
	LTE
	GTE
	INSIDE
	OUTSIDE
	BETWEEN
	COUNT
	RANGE
	LIMIT
	SORT
	VALUES
	VALUEMAP
	KEYS
	SUM
	ASC
	DESC
	IPV4RANGE
	SUBGRAPH
	FOREVER
	NOW
	AS
	SELECT

	TRUE
	FALSE

	// extensions token have to start after 1000
)
