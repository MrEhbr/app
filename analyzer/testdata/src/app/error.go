package app

var _ error = &Error{}

type Error struct {
	Err     error `json:"err"`
	Fields  map[string]interface{}
	Code    string `json:"code"`
	Message string `json:"message"`
	Op      string `json:"op"`
}

func (e Error) Error() string { return "test" }

func OpError(_ string, err error) error { return err }
