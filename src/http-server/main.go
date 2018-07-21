package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"context"
	"github.com/CX1ng/jaeger-example/src/common"
	"github.com/CX1ng/jaeger-example/src/jaeger_test"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	listenAddr = "127.0.0.1:8888"
	reportAddr = "127.0.0.1:5775"
	grpcAddr   = "127.0.0.1:8889"
)

func initTracing(cfg *config.Configuration) (io.Closer, error) {
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		return nil, err
	}
	opentracing.SetGlobalTracer(tracer)
	return closer, nil
}

func initJaegerCfg() (io.Closer, error) {
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  reportAddr,
		},
	}
	return cfg.InitGlobalTracer("Jaeger Http Server")
}

func tracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		closer, err := initJaegerCfg()
		if err != nil {

		}
		defer closer.Close()
		spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		var span opentracing.Span
		if err == nil {
			span = opentracing.GlobalTracer().StartSpan(r.URL.String(), opentracing.ChildOf(spanContext))
		} else {
			span = opentracing.GlobalTracer().StartSpan(r.URL.String())
		}
		defer span.Finish()
		span.SetTag("url", r.URL.String())
		span.SetTag("method", r.Method)
		span.LogKV("timestamp", time.Now().Format("2006/01/02 15:04:05"))
		md := metadata.New(nil)
		mdWriter := common.MdWriterReader{md}
		opentracing.GlobalTracer().Inject(span.Context(), opentracing.TextMap, mdWriter)
		ctx := metadata.NewOutgoingContext(r.Context(), md)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	closer, err := initJaegerCfg()
	if err != nil {
		panic(err)
	}
	defer closer.Close()
	http.Handle("/ping", tracingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("/pong"))
	})))
	http.Handle("/grpc", tracingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get msg
		if err := r.ParseForm(); err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		msg := "jaeger"
		if len(r.Form["msg"]) > 0 {
			msg = r.Form["msg"][0]
		}

		resp, err := grpcConn(r.Context(), msg)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte(resp))
	})))
	fmt.Printf("listen on %s\n", listenAddr)
	http.ListenAndServe(listenAddr, nil)
}

func grpcConn(ctx context.Context, msg string) (string, error) {
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		return "", err
	}
	client := jaeger_test.NewJaegerClient(conn)
	resp, err := client.SendMsg(ctx, &jaeger_test.Req{
		Msg: msg,
	})
	if err != nil {
		return "", err
	}
	return resp.Resp, nil
}
