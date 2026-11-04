package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"route256/loms/internal/app/server"
	"route256/loms/internal/pkg/kafka/producer"
	"route256/loms/internal/pkg/loms/repository/db_repository"
	"route256/loms/internal/pkg/middleware"
	"route256/loms/internal/pkg/util/logger"
	desc "route256/loms/pkg/api/loms/v1"
)

const (
	grpcPort  = 50051
	httpPort  = 8081
	dbConnEnv = "DB_CONN"
)

func main() {
	dbConnStr, exist := os.LookupEnv(dbConnEnv)
	if !exist || len(strings.TrimSpace(dbConnStr)) == 0 {
		panic("Environment variable DB_CONN is not present or empty!")
	}

	_, err := logger.New()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL("http://jaeger:4318"))
	if err != nil {
		logger.Panicw(ctx, "otel exporter error", "error", err.Error())
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("loms"),
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

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		logger.Panicw(ctx, "Failed to listen", "port", grpcPort, "err", err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.Panic,
			middleware.Logger,
			middleware.Validate,
		),
	)

	reflection.Register(grpcServer)

	dbConn, err := pgx.Connect(ctx, dbConnStr)
	if err != nil {
		logger.Panicw(ctx, fmt.Sprintf("Couldn't connect to DB! Check connection string: %s", dbConnStr), "error", err)
	}

	config := GetConfig(cliFlags)
	kafkaProducer, err := producer.NewSyncProducer(config.kafka)
	if err != nil {
		logger.Panicw(ctx, "Couldn't create kafka producer", "error", err)
	}
	defer kafkaProducer.Close()

	lomsUseCase := db_repository.NewLomsDao(
		dbConn,
		db_repository.NewQueryWrapper(db_repository.New(dbConn)),
		producer.NewKafkaSender(config.producer.TopicName, kafkaProducer),
	)
	controller := server.NewLomsServer(lomsUseCase)

	desc.RegisterLomsServer(grpcServer, controller)

	logger.Infow(ctx, fmt.Sprintf("server listening at %v", lis.Addr()))

	// grpcServer.Serve блокирующий, поэтому создаем в отдельной рутине
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			logger.Fatalw(ctx, "failed to serve gRPC", "error", err)
		}
	}()

	conn, err := grpc.NewClient(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatalw(ctx, "failed to create client", "error", err)
	}

	gwmux := runtime.NewServeMux()
	RegisterPprof(gwmux)

	if err = desc.RegisterLomsHandler(context.Background(), gwmux, conn); err != nil {
		logger.Fatalw(ctx, "failed to register gateway", "error", err)
	}

	gwServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: middleware.WithHTTPLoggingMiddleware(gwmux),
	}

	logger.Infow(ctx, fmt.Sprintf("Serving gRPC-Gateway on %s\n", gwServer.Addr))

	go func() {
		if err = gwServer.ListenAndServe(); err != nil {
			logger.Fatalw(ctx, "failed to serve gRPC-Gateway", "error", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Infow(ctx, fmt.Sprintf("Received signal \"%v\", shutting down...", sig))

	timeCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	grpcServer.GracefulStop()

	if err := gwServer.Shutdown(timeCtx); err != nil {
		logger.Fatalw(ctx, "failed to gracefully shutdown server", "error", err)
	}

	if err := dbConn.Close(timeCtx); err != nil {
		logger.Fatalw(ctx, "failed to close DB connection", "error", err)
	}
}

func RegisterPprof(mux *runtime.ServeMux) {
	mux.HandlePath(
		"GET",
		"/debug/pprof/",
		func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			pprof.Index(w, r)
		})
	mux.HandlePath(
		"GET",
		"/debug/pprof/cmdline",
		func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			pprof.Cmdline(w, r)
		})
	mux.HandlePath(
		"GET",
		"/debug/pprof/profile",
		func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			pprof.Profile(w, r)
		})
	mux.HandlePath(
		"GET",
		"/debug/pprof/symbol",
		func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			pprof.Symbol(w, r)
		})
	mux.HandlePath(
		"GET",
		"/debug/pprof/trace",
		func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			pprof.Trace(w, r)
		})
}
