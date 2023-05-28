package zerolog

import (
	"errors"

	"github.com/MrEhbr/app"
	"github.com/rs/zerolog"
)

type err struct {
	origin error
}

func ErrorMarshaler(e error) interface{} {
	if e == nil {
		return nil
	}

	return err{origin: e}
}

func (e err) MarshalZerologObject(event *zerolog.Event) {
	if !errors.Is(e.origin, &app.Error{}) {
		event.Str("msg", e.origin.Error())
		return
	}

	event.Str("msg", app.ErrorMessageDefault(e.origin, ""))
	event.Str("code", app.ErrorCode(e.origin).String())
	event.Strs("trace", app.ErrorTrace(e.origin))
	if fields := app.ErrorFields(e.origin); len(fields) > 0 {
		event.Dict("fields", zerolog.Dict().Fields(fields))
	}
}
