package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/guiferpa/gody/v2"
	"go.opentelemetry.io/otel"
	"route256/cart/internal/pkg/metrics"
	"route256/cart/internal/pkg/middleware"
	"route256/cart/internal/pkg/util/logger"
)

type checkoutCartRequest struct {
	UserID int64 `validate:"min=1"`
}

type checkoutCartResponse struct {
	OrderID int64
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

	var parseErr error = nil
	requestData := &checkoutCartRequest{}

	requestData.UserID, parseErr = parseInt64(request.PathValue("user_id"))
	if parseErr != nil {
		setInternalServerError(writer, "Parse query error")
		return
	}

	// Validation
	if valid, err := gody.DefaultValidate(*requestData, nil); err != nil || !valid {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	orderID, serviceErr := s.cartService.CheckoutCart(ctx, requestData.UserID)
	if serviceErr != nil {
		setInternalServerError(writer, "Service error")
		return
	}

	response := checkoutCartResponse{}
	response.OrderID = orderID

	buf, err := json.Marshal(&response)
	if err != nil {
		setInternalServerError(writer, "Json marshal error")
		return
	}

	writer.WriteHeader(http.StatusOK)

	if _, err = writer.Write(buf); err != nil {
		logger.Errorw(ctx, "Writer error", "error", err)
	}
}
