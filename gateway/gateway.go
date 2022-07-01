package gateway

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"

	"github.com/pascallin/go-kit-application/config"
	"github.com/pascallin/go-kit-application/pkg"
)

func main() {
	logger := pkg.GetLogger()

	r := mux.NewRouter()

	tracer := stdopentracing.GlobalTracer() // no-op
	zipkinTracer, _ := stdzipkin.NewTracer(nil, stdzipkin.WithNoopTracer(true))
	discoveryClient, err := pkg.NewKitDiscoverClient()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	registerAddsvc(ctx, r, tracer, zipkinTracer, discoveryClient.Client)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	httpAddr := fmt.Sprintf(":%d", config.GetGatewayConfig().HttpPort)
	// HTTP transport.
	go func() {
		logger.Log("transport", "HTTP", "addr", httpAddr)
		if err := http.ListenAndServe(httpAddr, r); err != nil {
			logger.Log(err.Error())
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	logger.Log("gateway", "exiting")
}
