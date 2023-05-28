package zerolog

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/MrEhbr/app"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/rs/zerolog"
)

func ExampleNewErrorMetricsWriter() {
	originalErrorMarshalFunc := zerolog.ErrorMarshalFunc
	defer func() {
		zerolog.ErrorMarshalFunc = originalErrorMarshalFunc
	}()

	zerolog.ErrorMarshalFunc = ErrorMarshaler

	registry := prometheus.NewRegistry()
	metricsWriter := NewErrorMetricsWriter(registry, "error.code", "errors")
	multi := zerolog.MultiLevelWriter(metricsWriter, os.Stdout)

	log := zerolog.New(multi)

	log.Error().Err(nil).Msg("nil error")
	log.Error().Err(errors.New("foo")).Msg("std error")
	log.Error().Err(app.ErrorWithCode(errors.New("foo"), app.ETEST)).Msg("wrapped std error")
	// Output: {"level":"error","message":"nil error"}
	// {"level":"error","error":{"msg":"foo"},"message":"std error"}
	// {"level":"error","error":{"msg":"foo","code":"test_error_code","trace":["github.com/MrEhbr/app/log/zerolog.ExampleNewErrorMetricsWriter"]},"message":"wrapped std error"}
}

func Test_metricsWriter_Write(t *testing.T) {
	const (
		metricName = "errors"
		errorKey   = "error"
	)
	originalErrorMarshalFunc := zerolog.ErrorMarshalFunc
	defer func() {
		zerolog.ErrorMarshalFunc = originalErrorMarshalFunc
	}()

	zerolog.ErrorMarshalFunc = ErrorMarshaler

	t.Run("metric inc", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		metricsWriter := NewErrorMetricsWriter(registry, "error.code", "errors")
		multi := zerolog.MultiLevelWriter(metricsWriter, io.Discard)

		log := zerolog.New(multi)

		log.Error().Err(&app.Error{Code: app.ETEST}).Msg("test")
		log.Error().Err(&app.Error{Code: app.ENOTFOUND}).Msg("test")
		sublogger := log.With().
			Str("foo", "bar").
			Logger()
		sublogger.Error().Err(&app.Error{Code: app.ETEST}).Msg("test")

		// The last \n at the end of this string is important
		expected := strings.NewReader(`
# HELP errors Number of errors grouped by code
# TYPE errors counter
errors{code="not_found"} 1
errors{code="test_error_code"} 2
`)
		if err := testutil.GatherAndCompare(registry, expected, "errors"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("not app.Error", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		metricsWriter := NewErrorMetricsWriter(registry, "error.code", "errors")
		multi := zerolog.MultiLevelWriter(metricsWriter, io.Discard)

		log := zerolog.New(multi)
		log.Error().Err(errors.New("test")).Msg("test")

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

	t.Run("don't send metric if level >= Error", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		metricsWriter := NewErrorMetricsWriter(registry, "error.code", "errors")
		multi := zerolog.MultiLevelWriter(metricsWriter, io.Discard)

		log := zerolog.New(multi)
		log.Info().Err(errors.New("test")).Msg("test")

		n, err := testutil.GatherAndCount(registry, "errors")
		if err != nil {
			t.Fatal(err)
		}

		if n != 0 {
			t.Fatalf("expected no metrics gathered, got: %d", n)
		}
	})
}

func Benchmark_metricsWriter_Write(b *testing.B) {
	registry := prometheus.NewRegistry()
	metricsWriter := NewErrorMetricsWriter(registry, "error.code", "errors")
	log := zerolog.New(metricsWriter)
	for i := 0; i < b.N; i++ {
		log.Error().Err(errors.New("test")).Msg("test")
	}
}
