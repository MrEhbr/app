package zap

import (
	"errors"

	"github.com/MrEhbr/app"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type err struct {
	origin error
}

func (e err) Error() string {
	return e.origin.Error()
}

func (e err) ErrorCode() string {
	return app.ErrorCode(e.origin).String()
}

type errFields map[string]interface{}

func (f errFields) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for k, v := range f {
		zap.Any(k, v).AddTo(enc)
	}
	return nil
}

type stringsArr []string

func (s stringsArr) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, v := range s {
		enc.AppendString(v)
	}
	return nil
}

func (e err) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if !errors.Is(e.origin, &app.Error{}) {
		enc.AddString("msg", e.origin.Error())
		return nil
	}

	enc.AddString("msg", app.ErrorMessageDefault(e.origin, ""))
	enc.AddString("code", app.ErrorCode(e.origin).String())
	enc.AddArray("trace", stringsArr(app.ErrorTrace(e.origin)))
	if fields := app.ErrorFields(e.origin); len(fields) > 0 {
		if err := enc.AddObject("fields", errFields(fields)); err != nil {
			return err
		}
	}

	return nil
}

func NamedError(key string, e error) zap.Field {
	if e == nil {
		return zap.Skip()
	}
	return zap.Field{Key: key, Type: zapcore.ObjectMarshalerType, Interface: err{origin: e}}
}

func Error(err error) zap.Field {
	return NamedError("error", err)
}
