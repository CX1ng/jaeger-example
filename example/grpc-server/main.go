package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"

	. "github.com/CX1ng/jaeger-example"
	"github.com/CX1ng/jaeger-example/example/jaeger_test"
	"github.com/uber/jaeger-client-go/config"
)

const (
	localAddr  = "127.0.0.1:8889"
	reportAddr = "127.0.0.1:5775"
)

func main() {
	sampleCfg := &config.SamplerConfig{Type: "const", Param: 1}
	reporterCfg := &config.ReporterConfig{BufferFlushInterval: 1 * time.Second, LogSpans: true, LocalAgentHostPort: reportAddr}
	tracer, err := InitTracerWithJaegerCfg("grpc-server", sampleCfg, reporterCfg)
	if err != nil {
		panic(err)
	}
	defer tracer.Close()
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		panic(err)
	}
	var opts []grpc.ServerOption
	// 拦截器
	var interceptor grpc.UnaryServerInterceptor
	interceptor = tracer.UnaryServerInterceptor()
	opts = append(opts, grpc.UnaryInterceptor(interceptor))
	server := grpc.NewServer(opts...)
	jaeger_test.RegisterJaegerServer(server, &Receiver{})
	fmt.Printf("grpc listen on %s\n", localAddr)
	if err := server.Serve(listener); err != nil {
		panic(err)
	}
}

type Receiver struct{}

func (r *Receiver) SendMsg(ctx context.Context, msg *jaeger_test.Req) (*jaeger_test.Resp, error) {
	req := msg.Msg
	reply := "hello " + req

	resp := &jaeger_test.Resp{
		Resp: reply,
	}
	return resp, nil
}
