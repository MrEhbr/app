package operation

import (
	"app"
)

func errorNotLast() (error, int) { // want "error must be last"
	return nil, 0
}

func valid() error {
	const op = "valid"
	return app.Error{Op: op, Message: "test"}
}

func noOpConst() (int, error) {
	return 0, app.Error{Op: "foo", Message: "test"} // want "const value must be used"
}

func invalidOpValue() error {
	const op = "invalid_OpValue" // want "operation must be `invalidOpValue` not `invalid_OpValue`"
	return app.Error{Op: op, Message: "bar"}
}

func wrongOpConstName() error {
	const operation = "wrongOpConstName" // want "operation constant must be named `op` not `operation`"
	return app.Error{Op: operation, Message: "bar"}
}

func returnsOpError() error {
	const op = "returnsOpError"
	return app.OpError(op, nil)
}

func returnsFunc() (int, error) {
	return noOpConst()
}
