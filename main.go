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

func initTracing(cfg *config.Configuration) (opentracing.Span,io.Closer,error) {
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger),)
	if err != nil {
		return nil,nil,err
	}
	return  tracer.StartSpan("http-server"), closer, nil
}

func initJaegerCfg() (*config.Configuration,error){
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
	return &cfg, nil
}

func main() {
	cfg, err := initJaegerCfg()
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		span,closer,err := initTracing(cfg)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		defer closer.Close()
		defer span.Finish()
		span.SetTag("method", r.Method)
		span.SetTag("url",r.URL)
		span.SetTag("timestamp", time.Now().Format("2006/01/02 15:04:05"))
		w.Write([]byte("pong"))
	})
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request){

		span,closer,err := initTracing(cfg)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		defer closer.Close()
		defer span.Finish()
		span.SetTag("url",r.URL)
		w.Write([]byte("test"))
	})
	fmt.Printf("listen on %s\n", listenAddr)
	http.ListenAndServe(listenAddr, nil)
}
