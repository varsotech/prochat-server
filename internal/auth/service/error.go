package service

type Error struct {
	// ExternalMessage is a message that can be relayed to the end user
	ExternalMessage string

	// HTTPCode is the http code that can be associated with the error
	HTTPCode int
}

func (e Error) Error() string {
	return e.ExternalMessage
}
