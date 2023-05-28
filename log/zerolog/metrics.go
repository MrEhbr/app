package zerolog

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/valyala/fastjson"
)

type metricsWriter struct {
	counter  *prometheus.CounterVec
	codePath []string
}

func NewErrorMetricsWriter(registry prometheus.Registerer, codePath, metricName string) *metricsWriter {
	return &metricsWriter{
		codePath: strings.Split(codePath, "."),
		counter: promauto.With(registry).NewCounterVec(prometheus.CounterOpts{
			Name: metricName,
			Help: "Number of errors grouped by code",
		}, []string{"code"}),
	}
}

func (w metricsWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	if level < zerolog.ErrorLevel {
		return len(p), nil
	}

	return w.Write(p)
}

func (w metricsWriter) Write(p []byte) (n int, err error) {
	code := fastjson.GetString(p, w.codePath...)
	if code == "" {
		code = "none"
	}

	w.counter.With(prometheus.Labels{"code": code}).Inc()
	return len(p), nil
}
