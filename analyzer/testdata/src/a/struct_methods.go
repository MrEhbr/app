package operation

import (
	"app"
)

type testStruct struct{}

func (testStruct) errorNotLast() (error, int) { // want "error must be last"
	return nil, 0
}

func (testStruct) valid() error {
	const op = "testStruct.valid"
	return app.Error{Op: op, Message: "test"}
}

func (testStruct) noOpConst() (int, error) {
	return 0, app.Error{Op: "foo", Message: "test"} // want "const value must be used"
}

func (testStruct) invalidOpValue() error {
	const op = "test_struct.invalidOpValue" // want "operation must be `testStruct.invalidOpValue` not `test_struct.invalidOpValue`"
	return app.Error{Op: op, Message: "bar"}
}

func (testStruct) wrongOpConstName() error {
	const operation = "testStruct.wrongOpConstName" // want "operation constant must be named `op` not `operation`"
	return app.Error{Op: operation, Message: "bar"}
}

func (testStruct) returnsOpError() error {
	const op = "testStruct.returnsOpError"
	return app.OpError(op, nil)
}

func (t testStruct) returnsFunc() (int, error) {
	return t.noOpConst()
}
