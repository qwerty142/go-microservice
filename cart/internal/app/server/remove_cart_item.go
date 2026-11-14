package server

import (
	"net/http"
	"time"

	"github.com/guiferpa/gody/v2"
	"go.opentelemetry.io/otel"
	"route256/cart/internal/pkg/metrics"
	"route256/cart/internal/pkg/middleware"
	"route256/cart/internal/pkg/util/logger"
)

type removeItemRequest struct {
	UserID int64 `validate:"min=1"`
	SkuID  int64 `validate:"min=1"`
}

// RemoveCartItem handles removing item from user's cart.
func (s *CartServer) RemoveCartItem(writerOrigin http.ResponseWriter, request *http.Request) {
	writer := middleware.NewResponseWriterWrapper(writerOrigin)
	defer func(startTime time.Time) {
		metrics.SaveRequestDuration(startTime, removeCartItemEndpoint, writer.GetStatusCode())
	}(time.Now())

	ctx, span := otel.Tracer(TracerName).Start(request.Context(), "removeCartItem")
	defer span.End()
	ctx = logger.ContextWithSpanValues(ctx, span)

	writer.Header().Set("Content-Type", "application/json")

	var parseErr error
	requestData := &removeItemRequest{}

	if requestData.UserID, parseErr = parseInt64(request.PathValue("user_id")); parseErr != nil {
		setBadRequestError(ctx, writer, "Parse query error")
		return
	}

	if requestData.SkuID, parseErr = parseInt64(request.PathValue("sku_id")); parseErr != nil {
		setBadRequestError(ctx, writer, "Parse query error")
		return
	}

	// Validation
	if valid, err := gody.DefaultValidate(*requestData, nil); err != nil || !valid {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if serviceErr := s.cartService.RemoveItemFromCart(ctx, requestData.UserID, requestData.SkuID); serviceErr != nil {
		setInternalServerError(ctx, writer, "Service error")
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
