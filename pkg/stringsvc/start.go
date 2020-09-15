package stringsvc

import (
	"context"
	"flag"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/pascallin/go-micro-services/common/register"
	"net/http"
	"os"

	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	httptransport "github.com/go-kit/kit/transport/http"
)

func StartStringSVCService() {
	var (
		listen = flag.String("listen", ":8091", "HTTP listen address")
		proxy  = flag.String("proxy", "", "Optional comma-separated list of URLs to proxy uppercase requests")
	)
	flag.Parse()

	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "listen", *listen, "caller", log.DefaultCaller)

	fieldKeys := []string{"method", "error"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)
	countResult := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "count_result",
		Help:      "The result of each count method.",
	}, []string{}) // no fields here

	var svc StringService
	svc = stringService{}
	svc = proxyingMiddleware(context.Background(), *proxy, logger)(svc)
	svc = loggingMiddleware{logger, svc}
	svc = instrumentingMiddleware{requestCount, requestLatency, countResult, svc}

	uppercaseHandler := httptransport.NewServer(
		makeUppercaseEndpoint(svc),
		DecodeUppercaseRequest,
		EncodeResponse,
	)
	countHandler := httptransport.NewServer(
		makeCountEndpoint(svc),
		DecodeCountRequest,
		EncodeResponse,
	)

	http.Handle("/uppercase", uppercaseHandler)
	http.Handle("/count", countHandler)
	http.Handle("/metrics", promhttp.Handler())

	ctrl, err := register.ConnConsul("http://localhost:8500")
	if err != nil {
		fmt.Errorf("register error")
	}
	ctrl.Register(&api.AgentServiceRegistration{
		Kind:              "HTTP",
		ID:                "addstring",
		Name:              "addstring",
		Tags:              []string{},
		Port:              8091,
		Address:           "127.0.0.1",
		EnableTagOverride: false,
		Meta:              map[string]string{},
		Weights: &api.AgentWeights{
			Passing: 10,
			Warning: 1,
		},
		//Check:             &api.AgentServiceCheck{
		//	Interval:                       "10s",
		//	Timeout:                        "5s",
		//	HTTP:                           "http://192.168.10.106:666/health",
		//	Method:                         "GET",
		//},})
	})

	logger.Log("msg", "HTTP", "addr", *listen)
	logger.Log("err", http.ListenAndServe(*listen, nil))
}
