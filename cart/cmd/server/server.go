package main

import (
	"context"
	"embed"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	loms_adapter "route256/cart/internal/adapters/loms"
	"route256/cart/internal/adapters/products"
	"route256/cart/internal/app/server"
	"route256/cart/internal/pkg/carts/repository"
	"route256/cart/internal/pkg/carts/service"
	"route256/cart/internal/pkg/middleware"
	"route256/cart/internal/pkg/util/logger"
	"route256/cart/pkg/api/loms/v1"
)

//go:embed config.json
var configJsonFile embed.FS

type Config struct {
	ServerConfig struct {
		HttpUrl   string `json:"http_url"`
		JaegerUrl string `json:"jaeger_url"`
	} `json:"server"`

	LomsServiceConfig struct {
		GrpcUrl string `json:"grpc_url"`
	} `json:"loms_service"`

	ProductServiceConfig struct {
		Token string `json:"token"`
		Retry struct {
			MaxCount     int   `json:"max_count"`
			DelaySeconds int   `json:"delay_seconds"`
			OnStatuses   []int `json:"on_statuses"`
		} `json:"retry"`
		Rps int `json:"rps"`
	} `json:"product_service"`
}

func main() {
	_, err := logger.New()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	configFileBytes, err := configJsonFile.ReadFile("config.json")
	if err != nil {
		logger.Fatalw(ctx, "opening config file", "error", err.Error())
	}

	var config Config
	if err = json.Unmarshal(configFileBytes, &config); err != nil {
		logger.Fatalw(ctx, "parsing config file", "error", err.Error())
	}

	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(config.ServerConfig.JaegerUrl))
	if err != nil {
		logger.Panicw(ctx, "otel exporter error", "error", err.Error())
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("cart"),
			semconv.DeploymentEnvironment("development"),
		),
	)
	if err != nil {
		logger.Panicw(ctx, "creating resource return error")
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)
	defer func() {
		traceProvider.Shutdown(ctx)
	}()

	otel.SetTracerProvider(traceProvider)

	cartRepository := repository.NewInMemoryCartRepository()
	productProvider := products.NewProductClientHTTP(
		config.ProductServiceConfig.Token,
		&http.Client{
			Transport: middleware.NewRetryableTransport(
				otelhttp.NewTransport(http.DefaultTransport),
				config.ProductServiceConfig.Retry.MaxCount,
				time.Duration(config.ProductServiceConfig.Retry.DelaySeconds)*time.Second,
				config.ProductServiceConfig.Retry.OnStatuses,
			),
		},
	)

	conn, err := grpc.NewClient(
		config.LomsServiceConfig.GrpcUrl,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		logger.Panicw(ctx, "failed creating gRPC client", "error", err.Error())
	}
	defer conn.Close()
	lomsClient := loms_adapter.NewClientWrapper(loms.NewLomsClient(conn))

	cartService := service.NewCartService(
		cartRepository,
		productProvider,
		lomsClient,
		config.ProductServiceConfig.Rps,
	)
	cartServer := server.NewCartServer(cartService)

	mux := cartServer.RegisterCartEndpoints(&http.ServeMux{})
	mux = cartServer.RegisterPprof(mux)
	handler := middleware.NewLogger(mux)

	httpServer := &http.Server{
		Addr:              config.ServerConfig.HttpUrl,
		ReadHeaderTimeout: 1 * time.Minute,
		Handler:           handler,
	}

	logger.Infow(ctx, "Serving on %s\n", httpServer.Addr)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			logger.Fatalw(ctx, "http server error", "error", err.Error())
		}
	}()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Infow(ctx, "Received signal \"%v\", shutting down...", sig)

	timeCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(timeCtx); err != nil {
		logger.Panicw(ctx, "http server shutdown error", "error", err.Error())
	}
}
