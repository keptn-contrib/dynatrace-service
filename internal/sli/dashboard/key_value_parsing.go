package dashboard

import "fmt"

type duplicateKeyError struct {
	key string
}

func (err *duplicateKeyError) Error() string {
	return fmt.Sprintf("duplicate key '%s' in SLO definition", err.key)
}
