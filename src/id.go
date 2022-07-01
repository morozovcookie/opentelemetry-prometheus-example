package otelexample

import (
	"fmt"
)

var _ fmt.Stringer = (*ID)(nil)

// ID describes the unique identifier.
type ID string

// The String method is used to print values passed as an operand
// to any format that accepts a string or to an unformatted printer
// such as Print.
func (id ID) String() string {
	return string(id)
}

// EmptyID is the constant for the identifier with empty value.
const EmptyID = ID("")
