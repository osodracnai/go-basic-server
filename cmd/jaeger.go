package cmd

import (
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
	jaegerPrometheus "github.com/uber/jaeger-lib/metrics/prometheus"
	"io"
	"time"
)

func (j *Jaeger) Flags(flags *pflag.FlagSet) {
	flags.String("jaeger.name", j.Name, "Service name used in jaeger")
	flags.Bool("jaeger.log-error", j.LogError, "Indicates whether to log jaeger error")
	flags.Bool("jaeger.log-info", j.LogInfo, "Indicates whether to log jaeger information")
}

func (j *Jaeger) Config() (io.Closer, error) {

	defcfg := jaegercfg.Configuration{
		ServiceName: j.Name,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}

	cfg, err := defcfg.FromEnv()
	if err != nil {
		return nil, err
	}

	jLogger := &stdLogger{
		LogError: j.LogError,
		LogInfo:  j.LogInfo,
	}
	jMetricsFactory := jaegerPrometheus.New()
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()

	closer, err := cfg.InitGlobalTracer(
		j.Name,
		jaegercfg.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.ZipkinSharedRPCSpan(true),
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		return nil, err
	}
	return closer, nil
}

type stdLogger struct {
	LogError bool
	LogInfo  bool
}

func (l *stdLogger) Error(msg string) {
	if l.LogError {
		logrus.WithFields(logrus.Fields{"jaeger": true}).Error(msg)
	}
}

// Infof logs a message at info priority
func (l *stdLogger) Infof(msg string, args ...interface{}) {
	if l.LogInfo {
		logrus.WithFields(logrus.Fields{"jaeger": true}).Infof(msg, args...)
	}
}
