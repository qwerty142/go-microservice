package server

import (
	"io"
	"net/http"
	"time"

	"github.com/guiferpa/gody/v2"
	"go.opentelemetry.io/otel"
	"route256/cart/internal/pkg/metrics"
	"route256/cart/internal/pkg/middleware"
	"route256/cart/internal/pkg/util/logger"
)

type clearCartRequest struct {
	UserID int64 `validate:"min=1"`
}

// ClearCart handles clearing user's cart.
func (s *CartServer) ClearCart(writerOrigin http.ResponseWriter, request *http.Request) {
	writer := middleware.NewResponseWriterWrapper(writerOrigin)
	defer func(startTime time.Time) {
		metrics.SaveRequestDuration(startTime, clearCartEndpoint, writer.GetStatusCode())
	}(time.Now())

	ctx, span := otel.Tracer(TracerName).Start(request.Context(), "clearCart")
	defer span.End()
	ctx = logger.ContextWithSpanValues(ctx, span)

	writer.Header().Set("Content-Type", "application/json")

	var parseErr error = nil
	requestData := &clearCartRequest{}

	requestData.UserID, parseErr = parseInt64(request.PathValue("user_id"))
	if parseErr != nil {
		setInternalServerError(writer, "Parse query error")
		return
	}

	defer func(body io.ReadCloser) {
		if closeErr := body.Close(); closeErr != nil {
			logger.Errorw(ctx, "Error closing response body", "error", closeErr)
		}
	}(request.Body)

	// Validation
	if valid, err := gody.DefaultValidate(*requestData, nil); err != nil || !valid {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if serviceErr := s.cartService.ClearUserCart(ctx, requestData.UserID); serviceErr != nil {
		setInternalServerError(writer, "Service error")
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
