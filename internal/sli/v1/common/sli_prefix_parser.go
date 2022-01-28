package common

import (
	"errors"
	"strings"
)

const (
	prefixDelimiter = ";"
)

// SLIPrefixParser parses a query string with ;-delimited prefixes into pieces.
type SLIPrefixParser struct {
	sli   string
	count int
}

// NewSLIPrefixParser creates a SLIPrefixParser for the specified SLI string and count.
func NewSLIPrefixParser(sli string, count int) SLIPrefixParser {
	return SLIPrefixParser{
		sli:   strings.TrimSpace(sli),
		count: count,
	}
}

// Parse parses the SLI string into number of pieces specified by count. The last piece is the remainder of the input string.
// An error is returned if the input string contains too few pieces.
func (e SLIPrefixParser) Parse() (*SLIPieces, error) {
	if e.count < 1 {
		return nil, errors.New("must parse into at least one piece")
	}

	pieces := strings.SplitN(e.sli, prefixDelimiter, e.count)
	if len(pieces) < e.count {
		return nil, errors.New("incorrect prefix")
	}

	sliPieces := NewSLIPieces(pieces)
	return &sliPieces, nil
}

func ProducePrefixedSLI(prefix string, otherPieces ...string) string {
	pieces := append([]string{prefix}, otherPieces...)
	return strings.Join(pieces, prefixDelimiter)
}
