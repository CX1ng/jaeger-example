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
	requestAddr = "http://127.0.0.1:8888/grpc?msg=beijing"
)

func InitJaegerCfg() (io.Closer, error) {
	cfg := &config.Configuration{
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
	return cfg.InitGlobalTracer("Jaeger Http Test")
}

func main() {
	closer, err := InitJaegerCfg()
	if err != nil {
		panic(err)
	}
	defer closer.Close()
	span := opentracing.GlobalTracer().StartSpan("request")
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
