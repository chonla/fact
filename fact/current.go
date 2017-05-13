package fact

import (
	"time"
)

type currentFn func() []interface{}

// CurrentFact is something about current thing
type CurrentFact struct {
}

// CurrentFactMapper is mapping table to fact resolution
var CurrentFactMapper = map[string]currentFn{
	"เวลา": findNow,
}

// NewCurrentFact is to create a new current fact
func NewCurrentFact() *CurrentFact {
	return &CurrentFact{}
}

// Find something in current stuff
func (c *CurrentFact) Find(what string) []interface{} {
	if _, ok := CurrentFactMapper[what]; ok {
		return CurrentFactMapper[what]()
	}
	return []interface{}{}
}

func findNow() []interface{} {
	now := time.Now()

	return []interface{}{now}
}
