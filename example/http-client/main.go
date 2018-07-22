package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/CX1ng/jaeger-example"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
)

const (
	reportAddr  = "127.0.0.1:5775"
	requestAddr = "http://127.0.0.1:8888/grpc?msg=beijing"
)

func main() {
	sampleCfg := &config.SamplerConfig{Type: "const", Param: 1}
	reporterCfg := &config.ReporterConfig{BufferFlushInterval: 1 * time.Second, LogSpans: true, LocalAgentHostPort: reportAddr}
	tracer, err := InitTracerWithJaegerCfg("http-client", sampleCfg, reporterCfg)
	if err != nil {
		panic(err)
	}
	defer tracer.Close()
	header := http.Header{}
	span := tracer.Tracer.StartSpan("server-client")
	defer span.Finish()
	tracer.Tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(header))

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
