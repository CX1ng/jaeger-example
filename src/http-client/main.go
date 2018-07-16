package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
)

const (
	reportAddr  = "127.0.0.1:5775"
	requestAddr = "http://127.0.0.1:8888/ping"
)

func InitJaegerCfg() *config.Configuration {
	cfg := &config.Configuration{
		ServiceName: "Jaeger HTTP Test",
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
	return cfg
}

func InitTracer(cfg *config.Configuration) (opentracing.Span, io.Closer, error) {
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil, nil, err
	}
	opentracing.SetGlobalTracer(tracer)
	span := opentracing.StartSpan("http-client")
	return span, closer, err
}

func main() {
	cfg := InitJaegerCfg()
	span, closer, err := InitTracer(cfg)
	if err != nil {
		panic(err)
	}
	defer closer.Close()
	defer span.Finish()

	header := http.Header{}
	span.SetTag("method", "Get")
	opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(header))

	client := http.Client{}
	req, err := http.NewRequest("Get", requestAddr, nil)
	if err != nil {
		panic(err)
	}
	req.Header = header
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("resp:%s", content)
}
