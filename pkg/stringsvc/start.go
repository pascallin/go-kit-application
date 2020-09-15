package stringsvc

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	kitlog "github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/pascallin/go-micro-services/common/register"
)

func StartStringSVCService() {
	var wait time.Duration
	var (
		listen = flag.String("listen", ":8091", "HTTP listen address")
		proxy  = flag.String("proxy", "", "Optional comma-separated list of URLs to proxy uppercase requests")
	)
	flag.Parse()

	var logger kitlog.Logger
	logger = kitlog.NewLogfmtLogger(os.Stderr)
	logger = kitlog.With(logger, "listen", *listen, "caller", kitlog.DefaultCaller)

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

	router := mux.NewRouter()
	router.Handle("/uppercase", uppercaseHandler).Methods("POST")
	router.Handle("/count", countHandler).Methods("POST")
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	ctrl, err := register.ConnConsul("http://localhost:8500")
	if err != nil {
		fmt.Errorf("register error")
	}

	ctrl.Register(&api.AgentServiceRegistration{
		Kind:              "HTTP",
		ID:                "stringsvc",
		Name:              "stringsvc",
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

	srv := &http.Server{
		Addr:         *listen,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: router, // Pass our instance of gorilla/mux in.
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	logger.Log("msg", "HTTP", "addr", *listen)


	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	ctrl.UnRegister("stringsvc")
	os.Exit(0)
	//logger.Log("msg", "HTTP", "addr", *listen)
	//logger.Log("err", http.ListenAndServe(*listen, nil))
}
