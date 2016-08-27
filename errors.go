package oddsengine

// InvalidUnitError represents an error where an invalid unit has been entered
// for a particular game.
type InvalidUnitError struct {
	s string
}

// Error returns the string form of the error
func (i InvalidUnitError) Error() string {
	return i.s
}
