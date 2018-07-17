package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"google.golang.org/grpc"

	. "github.com/CX1ng/jaeger-example/src/common"
	"github.com/CX1ng/jaeger-example/src/jaeger_test"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
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
			LogSpans:            true,
			LocalAgentHostPort:  reportAddr,
			BufferFlushInterval: 1 * time.Second,
		},
	}
	return cfg.InitGlobalTracer("grpc-server")
}

func initTracer(cfg *config.Configuration) (io.Closer, error) {
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil, err
	}
	opentracing.SetGlobalTracer(tracer)
	return closer, nil
}

func main() {
	closer, err := initJaegerCfg()
	if err != nil {
		panic(err)
	}
	defer closer.Close()

	// TODO: 拦截器
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	jaeger_test.RegisterJaegerServer(server, &Receiver{})
	fmt.Printf("grpc listen on %s\n", localAddr)
	if err := server.Serve(listener); err != nil {
		panic(err)
	}
}

type Receiver struct{}

func (r *Receiver) SendMsg(ctx context.Context, msg *jaeger_test.Req) (*jaeger_test.Resp, error) {
	var span opentracing.Span
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	spanContext, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, MdWriterReader{md})
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		return &jaeger_test.Resp{}, err
	} else if err != nil && err == opentracing.ErrSpanContextNotFound {
		span = opentracing.GlobalTracer().StartSpan("grpc-server")
	} else {
		span = opentracing.GlobalTracer().StartSpan("grpc-server", opentracing.ChildOf(spanContext))
	}
	defer span.Finish()
	req := msg.Msg
	reply := "hello " + req
	span.SetTag("method", "SendMsg")
	span.SetTag("req", req)
	span.SetTag("resp", reply)

	resp := &jaeger_test.Resp{
		Resp: reply,
	}
	return resp, nil
}
