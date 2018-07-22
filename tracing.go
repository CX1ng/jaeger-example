package jaeger_example

import (
	"context"
	"errors"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"net/http"
	"time"
)

type Tracer struct {
	Closer io.Closer
	Tracer opentracing.Tracer
}

func InitTracerWithJaegerCfg(serverName string, samplerCfg *config.SamplerConfig, reporterCfg *config.ReporterConfig) (*Tracer, error) {
	cfg := config.Configuration{
		ServiceName: serverName,
		Sampler:     samplerCfg,
		Reporter:    reporterCfg,
	}
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil, err
	}
	return &Tracer{
		Closer: closer,
		Tracer: tracer,
	}, nil
}

func (t *Tracer) Close() {
	t.Closer.Close()
}

func (t *Tracer) SpanFromHttpHeader(r *http.Request, operatorName string) (opentracing.Span, error) {
	spanContext, err := t.Tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		return nil, err
	}
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		return nil, err
	} else if err != nil && err == opentracing.ErrSpanContextNotFound {
		return t.Tracer.StartSpan(operatorName), nil
	} else {
		return t.Tracer.StartSpan(operatorName, opentracing.ChildOf(spanContext)), nil
	}
}

func (t *Tracer) HTTPMiddleware(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next := http.HandlerFunc(f)
		span, err := t.SpanFromHttpHeader(r, r.URL.String())
		if err != nil {
			next.ServeHTTP(w, r)
		}
		defer span.Finish()

		span.SetTag("URL", r.URL.String())
		span.SetTag("Method", r.Method)
		span.SetTag("Host", r.Host)
		span.LogKV("Timestamp", time.Now().Format("2016/01/02 15:04:05"))

		ctx := opentracing.ContextWithSpan(r.Context(), span)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (t *Tracer) SpanFromTextMap(ctx context.Context, operatorName string) (opentracing.Span, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("Not Found Md From Ctx")
	}
	reader := MdWriterReader{md}
	spanContext, err := t.Tracer.Extract(opentracing.TextMap, reader)
	if err != nil {
		return nil, err
	}
	var span opentracing.Span
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		return nil, err
	} else if err != nil && err == opentracing.ErrSpanContextNotFound {
		return t.Tracer.StartSpan(operatorName), nil
	} else {
		return t.Tracer.StartSpan(operatorName, opentracing.ChildOf(spanContext)), nil
	}
	return span, nil
}

func (t *Tracer) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		span, err := t.SpanFromTextMap(ctx, info.FullMethod)
		if err == nil {
			defer span.Finish()

			span.SetTag("FullMethod", info.FullMethod)
			span.SetTag("Server", info.Server)
			span.LogKV("Timestamp", time.Now().Format("2006/01/02 15:04:05"))
		}
		return handler(ctx, req)
	}
}

func (t *Tracer) InjectTextMap(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	textMap := MdWriterReader{md}
	span := opentracing.SpanFromContext(ctx)
	if err := t.Tracer.Inject(span.Context(), opentracing.TextMap, textMap); err == nil {
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}
