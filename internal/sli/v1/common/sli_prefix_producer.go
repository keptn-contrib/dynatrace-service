package common

import (
	"strings"
)

// SLIPrefixProducer build a SLI prefix string from SLIPieces.
type SLIPrefixProducer struct {
	pieces SLIPieces
}

// NewSLIPrefixProducer creates a new SLIPrefixProducer based on the specified pieces.
func NewSLIPrefixProducer(pieces SLIPieces) SLIPrefixProducer {
	return SLIPrefixProducer{
		pieces: pieces,
	}
}

// Produce produces a SLI prefix string based on the pieces.
func (b SLIPrefixProducer) Produce() string {
	p := make([]string, 0, b.pieces.Count())
	for i := 0; i < b.pieces.Count(); i++ {
		p = append(p, b.pieces.Get(i))
	}
	return strings.Join(p, prefixDelimiter)
}
