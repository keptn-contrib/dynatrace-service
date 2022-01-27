package common

import "errors"

// SLIPieces represents the parsed pieces of an SLI, typically created from a ;-delimited string.
type SLIPieces struct {
	pieces []string
}

// NewSLIPieces creates a new SLIPieces from the specified pieces.
func NewSLIPieces(pieces []string) SLIPieces {
	return SLIPieces{
		pieces: pieces,
	}
}

// Get gets the indexed piece.
func (p *SLIPieces) Get(index int) (string, error) {
	if (index < 0) || (index >= len(p.pieces)) {
		return "", errors.New("piece index of out of bounds")
	}

	return p.pieces[index], nil
}

// Count returns the number of pieces.
func (p *SLIPieces) Count() int {
	return len(p.pieces)
}
