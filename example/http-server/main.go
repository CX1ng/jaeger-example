package main

import (
	"fmt"
	"net/http"
	"time"

	"context"
	. "github.com/CX1ng/jaeger-example"
	"github.com/CX1ng/jaeger-example/example/jaeger_test"
	"github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
)

const (
	listenAddr = "127.0.0.1:8888"
	reportAddr = "127.0.0.1:5775"
	grpcAddr   = "127.0.0.1:8889"
)

func main() {
	sampleCfg := &config.SamplerConfig{Type: "const", Param: 1}
	reportCfg := &config.ReporterConfig{BufferFlushInterval: 1 * time.Second, LogSpans: true, LocalAgentHostPort: reportAddr}
	tracer, err := InitTracerWithJaegerCfg("server-http", sampleCfg, reportCfg)
	if err != nil {
		panic(err)
	}
	defer tracer.Close()

	http.Handle("/ping", tracer.HTTPMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("/pong"))
	}))
	http.Handle("/grpc", tracer.HTTPMiddleware(func(w http.ResponseWriter, r *http.Request) {
		ctx := tracer.InjectTextMap(r.Context())
		// get msg
		if err := r.ParseForm(); err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		msg := "jaeger"
		if len(r.Form["msg"]) > 0 {
			msg = r.Form["msg"][0]
		}

		resp, err := grpcConn(ctx, msg)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte(resp))
	}))
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
