package oddsengine

type InvalidPieceError struct {
	s string
}

func (i InvalidPieceError) Error() string {
	return i.s
}
