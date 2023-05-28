package app

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIs(t *testing.T) {
	tests := []error{
		&Error{},
		OpError("test", errors.New("test")),
		ErrorWithCode(errors.New("test"), ETEST),
		fmt.Errorf("%w", OpError("test", errors.New("test"))),
	}

	for i, tt := range tests {
		if !errors.Is(tt, &Error{}) {
			t.Fatalf("%d. %T: return false", i, tt)
		}
	}
}

func TestAs(t *testing.T) {
	testErr := &Error{Op: "test", Message: "foobar", Code: ETEST}
	tests := []struct {
		err  error
		want *Error
	}{
		{
			err:  testErr,
			want: testErr,
		},
		{
			err:  fmt.Errorf("%w", testErr),
			want: testErr,
		},
		{
			err:  fmt.Errorf("%w", fmt.Errorf("%w", fmt.Errorf("%w", testErr))),
			want: testErr,
		},
	}

	for i, tt := range tests {
		got := &Error{}
		if !errors.As(tt.err, &got) {
			t.Fatalf("%d. %T: return false", i, tt.err)
		}

		if got.Code != tt.want.Code {
			t.Fatalf("%d. Code want: %s, got: %s", i, tt.want.Code, got.Code)
		}

		if got.Message != tt.want.Message {
			t.Fatalf("%d. Message want: %s, got: %s", i, tt.want.Message, got.Message)
		}
		if got.Op != tt.want.Op {
			t.Fatalf("%d. Op want: %s, got: %s", i, tt.want.Op, got.Op)
		}
	}
}

func TestErrorCode(t *testing.T) {
	tests := []struct {
		err  error
		want errorCode
	}{
		{
			err:  &Error{Code: ETEST},
			want: ETEST,
		},
		{
			err:  fmt.Errorf("%w", &Error{Code: ETEST}),
			want: ETEST,
		},
		{
			err:  fmt.Errorf("%w", fmt.Errorf("%w", fmt.Errorf("%w", &Error{Code: ETEST}))),
			want: ETEST,
		},
		{
			err: &Error{
				Err: &Error{
					Code: ETEST,
				},
			},
			want: ETEST,
		},
		{
			err:  &Error{Err: &Error{}},
			want: EINTERNAL,
		},
		{
			err:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		got := ErrorCode(tt.err)
		if tt.want != got {
			t.Fatalf("ErrorCode want: %s, got: %s", tt.want, got)
		}

	}
}

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{
			err:  &Error{Message: "foo"},
			want: "foo",
		},
		{
			err:  fmt.Errorf("%w", &Error{Message: "foo"}),
			want: "foo",
		},
		{
			err:  fmt.Errorf("%w", fmt.Errorf("%w", fmt.Errorf("%w", &Error{Message: "foo"}))),
			want: "foo",
		},
		{
			err: &Error{
				Err: &Error{
					Message: "foo",
				},
			},
			want: "foo",
		},
		{
			err:  &Error{Err: &Error{}},
			want: DefaultErrorMessage,
		},
		{
			err:  &Error{Err: errors.New("foo")},
			want: DefaultErrorMessage,
		},
		{
			err:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		got := ErrorMessage(tt.err)
		if tt.want != got {
			t.Fatalf("ErrorMessage want: %s, got: %s", tt.want, got)
		}

	}
}

func TestErrorFields(t *testing.T) {
	tests := []struct {
		err  error
		want map[string]interface{}
	}{
		{
			err:  &Error{Message: "foo", Fields: map[string]interface{}{"foo": "bar"}},
			want: map[string]interface{}{"foo": "bar"},
		},
		{
			err:  fmt.Errorf("%w", &Error{Message: "foo", Fields: map[string]interface{}{"foo": "bar"}}),
			want: map[string]interface{}{"foo": "bar"},
		},
		{
			err:  fmt.Errorf("%w", fmt.Errorf("%w", fmt.Errorf("%w", &Error{Message: "foo", Fields: map[string]interface{}{"foo": "bar"}}))),
			want: map[string]interface{}{"foo": "bar"},
		},
		{
			err: &Error{
				Fields: map[string]interface{}{"foo": "bar"},
				Err: &Error{
					Message: "foo",
					Fields:  map[string]interface{}{"bar": "baz"},
				},
			},
			want: map[string]interface{}{"foo": "bar", "bar": "baz"},
		},
		{
			err: &Error{
				Fields: map[string]interface{}{"foo": "bar"},
				Err: fmt.Errorf("%w", &Error{
					Message: "foo",
					Fields:  map[string]interface{}{"bar": "baz"},
				}),
			},
			want: map[string]interface{}{"foo": "bar", "bar": "baz"},
		},
		{
			err:  &Error{Err: &Error{}},
			want: map[string]interface{}{},
		},
		{
			err:  nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		got := ErrorFields(tt.err)
		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf("ErrorFields() mismatch (-want +got):\n%s", diff)
		}
	}
}

func eq(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if w, ok := b[k]; !ok || v != w {
			return false
		}
	}

	return true
}

func TestErrorWithCode(t *testing.T) {
	type args struct {
		err  error
		code errorCode
	}
	// If Error.Op is undefined, caller function name will be used as Op
	// If the err is an Error && err.Code is defined, the err is wrapped and the code is applied;
	tests := []struct {
		name string
		want *Error
		args args
	}{
		{
			name: "err is a regular error",
			args: args{
				err:  errors.New("test"),
				code: ETEST,
			},
			want: &Error{Op: "github.com/MrEhbr/app.errWithCode", Code: ETEST, Err: errors.New("test")},
		},
		{
			name: "err is an Error && err.Code is undefined, op not empty",
			args: args{
				err:  &Error{Op: "test", Message: "test"},
				code: ETEST,
			},
			want: &Error{Op: "test", Code: ETEST, Message: "test"},
		},
		{
			name: "err is an Error && err.Code is undefined, op empty",
			args: args{
				err:  &Error{Message: "test"},
				code: ETEST,
			},
			want: &Error{Op: "github.com/MrEhbr/app.errWithCode", Code: ETEST, Message: "test"},
		},
		{
			name: "err is an Error && err.Code is defined",
			args: args{
				err:  &Error{Op: "test", Code: ENOTFOUND, Message: "test"},
				code: ETEST,
			},
			want: &Error{Op: "github.com/MrEhbr/app.errWithCode", Code: ETEST, Err: &Error{Op: "test", Code: ENOTFOUND, Message: "test"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errWithCode(tt.args.err, tt.args.code)
			got := &Error{}
			if !errors.As(err, &got) {
				t.Fatalf("%s. %T: return false", tt.name, tt.args.err)
			}

			if got.Code != tt.want.Code {
				t.Fatalf("%s. Code want: %s, got: %s", tt.name, tt.want.Code, got.Code)
			}

			if got.Message != tt.want.Message {
				t.Fatalf("%s. Message want: %s, got: %s", tt.name, tt.want.Message, got.Message)
			}

			if got.Op != tt.want.Op {
				t.Fatalf("%s. Op want: %s, got: %s", tt.name, tt.want.Op, got.Op)
			}
		})
	}
}

func errWithCode(err error, code errorCode) error {
	return ErrorWithCode(err, code)
}

func TestOpErrorOrNil(t *testing.T) {
	type args struct {
		op  string
		err error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "err is nil",
			args: args{
				op:  "test",
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "err not nil",
			args: args{
				op:  "test",
				err: errors.New("test"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := OpErrorOrNil(tt.args.op, tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("OpErrorOrNil() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func foo() error {
	panic("foo")
	return OpError("foo", errors.New("foo"))
}

func bar() error {
	return OpError("bar", foo())
}

func baz() error {
	return OpError("baz", bar())
}

func ExampleTrace() {
	const op = "ExampleRun"
	err := baz()
	fmt.Println(ErrorTrace(OpError(op, err)))
	// Output:
	// [ExampleRun baz bar foo]
}
