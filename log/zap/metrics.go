package zap

import (
	"github.com/MrEhbr/app"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ zapcore.Core = &metricCore{}

type metricCore struct {
	counter  *prometheus.CounterVec
	errorKey string
}

// NewErrorMetricsCore return core that will send metrics when log error
func NewErrorMetricsCore(registry prometheus.Registerer, metricName, errorKey string) zapcore.Core {
	return &metricCore{
		errorKey: errorKey,
		counter: promauto.With(registry).NewCounterVec(prometheus.CounterOpts{
			Name: metricName,
			Help: "Number of errors grouped by code",
		}, []string{"code"}),
	}
}

func (c *metricCore) Enabled(level zapcore.Level) bool {
	return zap.ErrorLevel.Enabled(level)
}

func (c *metricCore) Sync() error {
	return nil
}

func (c *metricCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}

	return ce
}

func (c *metricCore) With(fields []zapcore.Field) zapcore.Core {
	return c
}

func (c *metricCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	errField, ok := haveAppError(c.errorKey, fields)
	if !ok {
		c.counter.With(prometheus.Labels{"code": "none"}).Inc()
		return nil
	}

	errorCode := app.ErrorCode(errField.Interface.(error)).String()
	if v, ok := errField.Interface.(interface{ ErrorCode() string }); ok {
		errorCode = v.ErrorCode()
	}

	c.counter.With(prometheus.Labels{"code": errorCode}).Inc()

	return nil
}

func haveAppError(key string, fields []zapcore.Field) (zapcore.Field, bool) {
	for _, field := range fields {
		if field.Key == key {
			_, ok := field.Interface.(error)
			if !ok {
				return zapcore.Field{}, false
			}
			return field, true
		}
	}
	return zapcore.Field{}, false
}
