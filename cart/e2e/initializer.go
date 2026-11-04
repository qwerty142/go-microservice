package e2e

import (
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"route256/cart/internal/adapters/products"
	"route256/cart/internal/app/server"
	"route256/cart/internal/pkg/carts/repository"
	"route256/cart/internal/pkg/carts/service"
	"route256/cart/internal/pkg/middleware"
	"route256/cart/pkg/api/loms/v1"
)

const httpUrl = "http://localhost:8082"

func initServer() *http.ServeMux {
	cartRepository := repository.NewInMemoryCartRepository()
	productProvider := products.NewProductClientHTTP(
		"testtoken",
		&http.Client{
			Transport: middleware.NewRetryableTransport(http.DefaultTransport, 3, 1*time.Second, []int{420, 429}),
		},
	)

	conn, err := grpc.NewClient(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	lomsClient := loms.NewLomsClient(conn)

	cartService := service.NewCartService(cartRepository, productProvider, lomsClient)
	cartServer := server.NewCartServer(cartService)
	return cartServer.RegisterCartEndpoints(&http.ServeMux{})
}
