package transports

import (
	"context"
	"encoding/json"
	"net/http"

	kittransport "github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	kitlog "github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/swaggo/http-swagger"

	"github.com/pascallin/go-kit-application/middleware"
	_ "github.com/pascallin/go-kit-application/usersvc/docs"
	"github.com/pascallin/go-kit-application/usersvc/endpoints"
	"github.com/pascallin/go-kit-application/usersvc/services"
)

func MakeHandler(s services.Service, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(kittransport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(middleware.ErrorEncoder),
	}

	r := mux.NewRouter()

	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), //The url pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		// httpSwagger.DomID("#swagger-ui"),
	)).Methods(http.MethodGet)

	r.Handle("/user/v1/register", registerHandler(s, opts, logger)).Methods("POST")

	return r
}

// user register godoc
// @Summary user register
// @Schemes
// @Description create subscription
// @Tags user
// @Accept json
// @Produce json
// @security  ServiceApiKey
// @Success 200 {object} endpoints.RegisterRequest
// @Router /user/v1/register [post]
// @Param   data     body    endpoints.RegisterResponse     true        "data"
func registerHandler(s services.Service, opts []kithttp.ServerOption, logger kitlog.Logger) *kithttp.Server {
	return kithttp.NewServer(
		middleware.LoggingMiddleware(kitlog.With(logger, "method", "user register"))(endpoints.MakeRegisterEndpoint(s)),
		decodeRegisterRequest,
		encodeResponse,
		opts...,
	)
}

type errorer interface {
	error() error
}

func decodeRegisterRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	return endpoints.RegisterRequest{
		Username: body.Username,
		Password: body.Password,
		Nickname: body.Nickname,
	}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
