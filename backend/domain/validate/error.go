package validate

// Error is a rejected field or rule. HTTP maps it to 400.
type Error struct {
	Msg string
}

func (e *Error) Error() string {
	return e.Msg
}
