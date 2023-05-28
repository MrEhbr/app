package zap

import (
	"errors"

	"github.com/MrEhbr/app"
	"go.uber.org/zap"
)

func ExampleNew() {
	log := zap.NewExample()

	log.Info("nil error", Error(nil))
	log.Info("std error", Error(errors.New("foo")))
	log.Info("wrapped std error", Error(app.ErrorWithCode(errors.New("foo"), app.ETEST)))
	log.Info("error with fields", Error(&app.Error{Op: "test", Code: app.ETEST, Message: "foo", Fields: map[string]interface{}{"foo": "bar"}}))
	// Output: {"level":"info","msg":"nil error"}
	// {"level":"info","msg":"std error","error":{"msg":"foo"}}
	// {"level":"info","msg":"wrapped std error","error":{"msg":"foo","code":"test_error_code","trace":["github.com/MrEhbr/app/log/zap.ExampleNew"]}}
	// {"level":"info","msg":"error with fields","error":{"msg":"foo","code":"test_error_code","trace":["test"],"fields":{"foo":"bar"}}}
}
