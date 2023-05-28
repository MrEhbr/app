package zerolog

import (
	"errors"
	"os"

	"github.com/MrEhbr/app"
	"github.com/rs/zerolog"
)

func ExampleNew() {
	originalErrorMarshalFunc := zerolog.ErrorMarshalFunc
	defer func() {
		zerolog.ErrorMarshalFunc = originalErrorMarshalFunc
	}()

	zerolog.ErrorMarshalFunc = ErrorMarshaler

	log := zerolog.New(os.Stdout)

	log.Error().Err(nil).Msg("nil error")
	log.Error().Err(errors.New("foo")).Msg("std error")
	log.Error().Err(app.ErrorWithCode(errors.New("foo"), app.ETEST)).Msg("wrapped std error")
	log.Error().Err(&app.Error{Op: "test", Code: app.ETEST, Message: "foo", Fields: map[string]interface{}{"foo": "bar"}}).Msg("error with fields")

	// Output: {"level":"error","message":"nil error"}
	// {"level":"error","error":{"msg":"foo"},"message":"std error"}
	// {"level":"error","error":{"msg":"foo","code":"test_error_code","trace":["github.com/MrEhbr/app/log/zerolog.ExampleNew"]},"message":"wrapped std error"}
	// {"level":"error","error":{"msg":"foo","code":"test_error_code","trace":["test"],"fields":{"foo":"bar"}},"message":"error with fields"}
}
