package main

import (
	"fmt"
	"net/http"
	"time"
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go"
)

const (
	listenAddr = "127.0.0.1:8888"
	reportAddr = "127.0.0.1:5775"
)

func initTracing(cfg *config.Configuration) (io.Closer,error) {
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger),)
	if err != nil {
		return nil,err
	}
	opentracing.SetGlobalTracer(tracer)
	return  closer, nil
}

func initJaegerCfg() *config.Configuration{
	cfg := config.Configuration{
		ServiceName: "Jaeger Test",

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
	return &cfg
}

func main() {
	cfg := initJaegerCfg()
	closer,err := initTracing(cfg)
	if err != nil {
		panic(err)
	}
	defer closer.Close()
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		spanContext,err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		var span opentracing.Span
		if err == nil {
			span = opentracing.StartSpan("http-server-ping", opentracing.ChildOf(spanContext))
		}else{
			span = opentracing.GlobalTracer().StartSpan("http-server-ping")
		}
		defer span.Finish()
		span.SetTag("method", r.Method)
		span.SetTag("url",r.URL)
		span.SetTag("timestamp", time.Now().Format("2006/01/02 15:04:05"))
		w.Write([]byte("pong"))
	})
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request){
		var span opentracing.Span
		spanContext,err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		if spanContext != nil {
			span = opentracing.StartSpan("http-server-test", opentracing.ChildOf(spanContext))
		}else{
			span = opentracing.StartSpan("http-server-test")
		}
		defer span.Finish()
		span.SetTag("url",r.URL)
		w.Write([]byte("test"))
	})
	fmt.Printf("listen on %s\n", listenAddr)
	http.ListenAndServe(listenAddr, nil)
}
