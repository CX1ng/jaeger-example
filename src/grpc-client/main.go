package main

import (
	"context"
	"fmt"
	"io"
	"time"

	. "github.com/CX1ng/jaeger-example/src/common"
	"github.com/CX1ng/jaeger-example/src/jaeger_test"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	localAddr  = "127.0.0.1:8889"
	reportAddr = "127.0.0.1:5775"
)

func initJaegerCfg() (io.Closer, error) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			BufferFlushInterval: 1 * time.Second,
			LogSpans:            true,
			LocalAgentHostPort:  reportAddr,
		},
	}
	return cfg.InitGlobalTracer("grpc-client")
}

func main() {
	closer, err := initJaegerCfg()
	if err != nil {
		panic(err)
	}
	defer closer.Close()

	span := opentracing.GlobalTracer().StartSpan("grpc-client")
	defer span.Finish()
	span.SetTag("method", "grpc-client")
	span.SetTag("req", "jaeger")

	md := metadata.New(nil)
	mdWriter := MdWriterReader{md}
	err = opentracing.GlobalTracer().Inject(span.Context(), opentracing.TextMap, mdWriter)
	if err != nil {
		panic(err)
	}
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	conn, err := grpc.Dial(localAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := jaeger_test.NewJaegerClient(conn)
	req := &jaeger_test.Req{
		Msg: "jaeger",
	}
	resp, err := client.SendMsg(ctx, req)
	if err != nil {
		panic(err)
	}
	content := resp.Resp
	fmt.Printf("resp:%v\n", content)
}
