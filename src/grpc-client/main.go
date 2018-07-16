package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/CX1ng/jaeger-example/src/jaeger_test"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strings"
)

const (
	localAddr  = "127.0.0.1:8889"
	reportAddr = "127.0.0.1:5775"
)

func initJaegerCfg() *config.Configuration {
	cfg := &config.Configuration{
		ServiceName: "grpc-client",
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
	return cfg
}

func initTracer(cfg *config.Configuration) (io.Closer, error) {
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil, err
	}
	opentracing.SetGlobalTracer(tracer)
	return closer, nil
}

type MapWriterReader struct {
	metadata.MD
}

func (m MapWriterReader) Set(key, val string) {
	key = strings.ToLower(key)
	m.MD[key] = append(m.MD[key], val)
}

func (m MapWriterReader) ForeachKey(handler func(key, val string) error) error {
	for k, vs := range m.MD {
		for _, v := range vs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	cfg := initJaegerCfg()
	closer, err := initTracer(cfg)
	if err != nil {
		panic(err)
	}
	defer closer.Close()

	span := opentracing.StartSpan("grpc-client")
	defer span.Finish()
	span.SetTag("method", "grpc-client")
	span.SetTag("req", "jaeger")

	md := metadata.New(nil)
	mdWriter := MapWriterReader{md}
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
