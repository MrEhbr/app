package zap

import (
	"errors"
	"strings"
	"testing"

	"github.com/MrEhbr/app"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ExampleNewErrorMetricsCore() {
	log := zap.NewExample()
	registry := prometheus.NewRegistry()
	log = log.WithOptions(zap.WrapCore(func(origin zapcore.Core) zapcore.Core {
		return zapcore.NewTee(origin, NewErrorMetricsCore(registry, "errors", "error"))
	}),
	)

	log.Error("test")
	log.Error("test", Error(&app.Error{Code: app.ETEST, Message: "test"}))
	log.With(zap.String("logger", "sublogger")).Error("test", Error(&app.Error{Op: "test", Code: app.ETEST, Message: "test"}))
	// Output: {"level":"error","msg":"test"}
	// {"level":"error","msg":"test","error":{"msg":"test","code":"test_error_code","trace":[]}}
	// {"level":"error","msg":"test","logger":"sublogger","error":{"msg":"test","code":"test_error_code","trace":["test"]}}
}

func Test_metricCore_Write(t *testing.T) {
	const (
		metricName = "errors"
		errorKey   = "error"
	)
	t.Run("metric inc", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		log := zap.NewNop()
		log = log.WithOptions(zap.WrapCore(func(origin zapcore.Core) zapcore.Core {
			return zapcore.NewTee(origin, NewErrorMetricsCore(registry, "errors", "error"))
		}),
		)

		log.Error("test")
		log.Error("test", Error(&app.Error{Code: app.ETEST}))
		log.Error("test", Error(&app.Error{Code: app.ENOTFOUND}))
		log.With(zap.String("foo", "bar")).Error("test", Error(&app.Error{Code: app.ETEST}))

		// The last \n at the end of this string is important
		expected := strings.NewReader(`
	# HELP errors Number of errors grouped by code
	# TYPE errors counter
	errors{code="none"} 1
	errors{code="not_found"} 1
	errors{code="test_error_code"} 2
	`)
		if err := testutil.GatherAndCompare(registry, expected, "errors"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("invalid error field", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		log := zap.NewNop()
		log = log.WithOptions(zap.WrapCore(func(origin zapcore.Core) zapcore.Core {
			return zapcore.NewTee(origin, NewErrorMetricsCore(registry, "errors", "error"))
		}),
		)

		log.Error("test", zap.String(errorKey, "test"))

		// The last \n at the end of this string is important
		expected := strings.NewReader(`
	# HELP errors Number of errors grouped by code
	# TYPE errors counter
	errors{code="none"} 1
	`)
		if err := testutil.GatherAndCompare(registry, expected, "errors"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("don't send metric if level < Error", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		log := zap.NewNop()
		log = log.WithOptions(zap.WrapCore(func(origin zapcore.Core) zapcore.Core {
			return zapcore.NewTee(origin, NewErrorMetricsCore(registry, "errors", "error"))
		}),
		)

		log.Info("test", Error(errors.New("test")))

		n, err := testutil.GatherAndCount(registry, "errors")
		if err != nil {
			t.Fatal(err)
		}

		if n != 0 {
			t.Fatalf("expected no metrics gathered, got: %d", n)
		}
	})
}
