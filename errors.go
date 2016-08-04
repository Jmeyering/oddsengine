package oddsengine

// InvalidPieceError represents an error where an invalid piece has been entered
// for a particular game.
type InvalidPieceError struct {
	s string
}

// Error returns the string form of the error
func (i InvalidPieceError) Error() string {
	return i.s
}
