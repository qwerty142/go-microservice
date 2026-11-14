package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/guiferpa/gody/v2"
	"go.opentelemetry.io/otel"
	"route256/cart/internal/pkg/carts/service"
	"route256/cart/internal/pkg/metrics"
	"route256/cart/internal/pkg/middleware"
	"route256/cart/internal/pkg/util/logger"
)

type checkoutCartRequest struct {
	UserID int64 `validate:"min=1"`
}

type checkoutCartResponse struct {
	OrderID int64 `json:"order_id"`
}

// CheckoutCart handles order creation.
func (s *CartServer) CheckoutCart(writerOrigin http.ResponseWriter, request *http.Request) {
	writer := middleware.NewResponseWriterWrapper(writerOrigin)
	defer func(startTime time.Time) {
		metrics.SaveRequestDuration(startTime, checkoutCartEndpoint, writer.GetStatusCode())
	}(time.Now())

	ctx, span := otel.Tracer(TracerName).Start(request.Context(), "checkoutCart")
	defer span.End()
	ctx = logger.ContextWithSpanValues(ctx, span)

	writer.Header().Set("Content-Type", "application/json")

	var parseErr error
	requestData := &checkoutCartRequest{}

	if requestData.UserID, parseErr = parseInt64(request.PathValue("user_id")); parseErr != nil {
		setBadRequestError(ctx, writer, "Parse query error")
		return
	}

	// Validation
	if valid, err := gody.DefaultValidate(*requestData, nil); err != nil || !valid {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	orderID, serviceErr := s.cartService.CheckoutCart(ctx, requestData.UserID)
	if serviceErr != nil {
		if errors.Is(serviceErr, service.ErrCartNotFound) {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		setInternalServerError(ctx, writer, "Service error")
		return
	}

	buf, err := json.Marshal(&checkoutCartResponse{OrderID: orderID})
	if err != nil {
		setInternalServerError(ctx, writer, "Json marshal error")
		return
	}

	writer.WriteHeader(http.StatusOK)

	if _, err = writer.Write(buf); err != nil {
		logger.Errorw(ctx, "Writer error", "error", err)
	}
}
