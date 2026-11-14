package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/guiferpa/gody/v2"
	"go.opentelemetry.io/otel"
	"route256/cart/internal/adapters/products"
	"route256/cart/internal/pkg/carts/service"
	"route256/cart/internal/pkg/metrics"
	"route256/cart/internal/pkg/middleware"
	"route256/cart/internal/pkg/util/logger"
)

type addItemRequest struct {
	UserID int64  `validate:"min=1"`
	SkuID  int64  `validate:"min=1"`
	Count  uint16 `validate:"min=1" json:"count"`
}

// AddCartItem handles adding an item to the user's cart.
func (s *CartServer) AddCartItem(writerOrigin http.ResponseWriter, request *http.Request) {
	writer := middleware.NewResponseWriterWrapper(writerOrigin)
	defer func(startTime time.Time) {
		metrics.SaveRequestDuration(startTime, addCartItemEndpoint, writer.GetStatusCode())
	}(time.Now())

	ctx, span := otel.Tracer(TracerName).Start(request.Context(), "addCartItem")
	defer span.End()
	ctx = logger.ContextWithSpanValues(ctx, span)

	writer.Header().Set("Content-Type", "application/json")

	requestData := &addItemRequest{}

	var parseErr error
	if requestData.UserID, parseErr = parseInt64(request.PathValue("user_id")); parseErr != nil {
		setBadRequestError(ctx, writer, "Parse query error")
		return
	}

	if requestData.SkuID, parseErr = parseInt64(request.PathValue("sku_id")); parseErr != nil {
		setBadRequestError(ctx, writer, "Parse query error")
		return
	}

	defer func(body io.ReadCloser) {
		if closeErr := body.Close(); closeErr != nil {
			logger.Errorw(ctx, "Error closing request body", "error", closeErr)
		}
	}(request.Body)

	if err := json.NewDecoder(request.Body).Decode(requestData); err != nil {
		setBadRequestError(ctx, writer, "Parse body error")
		return
	}

	// Validation
	if valid, err := gody.DefaultValidate(*requestData, nil); err != nil || !valid {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	serviceErr := s.cartService.AddItemToCart(ctx, requestData.UserID, requestData.SkuID, requestData.Count)
	if serviceErr != nil {
		var requestErr *products.RequestError

		switch {
		case errors.As(serviceErr, &requestErr) && requestErr.NotFound(),
			errors.Is(serviceErr, service.ErrInsufficientStocks):
			writer.WriteHeader(http.StatusPreconditionFailed)
		default:
			setInternalServerError(ctx, writer, "Service error")
		}

		return
	}

	writer.WriteHeader(http.StatusOK)
}
