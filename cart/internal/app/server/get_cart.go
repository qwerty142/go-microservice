package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/guiferpa/gody/v2"
	"go.opentelemetry.io/otel"
	"route256/cart/internal/pkg/carts/model"
	"route256/cart/internal/pkg/carts/service"
	"route256/cart/internal/pkg/metrics"
	"route256/cart/internal/pkg/middleware"
	"route256/cart/internal/pkg/util/logger"
)

type getCartRequest struct {
	UserID int64 `validate:"min=1"`
}

type getCartResponse struct {
	Items      []model.Item `json:"items"`
	TotalPrice uint32       `json:"total_price"`
}

// GetCart handles retrieving user's cart.
func (s *CartServer) GetCart(writerOrigin http.ResponseWriter, request *http.Request) {
	writer := middleware.NewResponseWriterWrapper(writerOrigin)
	defer func(startTime time.Time) {
		metrics.SaveRequestDuration(startTime, getCartEndpoint, writer.GetStatusCode())
	}(time.Now())

	ctx, span := otel.Tracer(TracerName).Start(request.Context(), "getCart")
	defer span.End()
	ctx = logger.ContextWithSpanValues(ctx, span)

	writer.Header().Set("Content-Type", "application/json")

	var parseErr error
	requestData := &getCartRequest{}

	if requestData.UserID, parseErr = parseInt64(request.PathValue("user_id")); parseErr != nil {
		setBadRequestError(ctx, writer, "Parse query error")
		return
	}

	// Validation
	if valid, err := gody.DefaultValidate(*requestData, nil); err != nil || !valid {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	cart, serviceErr := s.cartService.GetUserCart(ctx, requestData.UserID)
	if serviceErr != nil {
		if errors.Is(serviceErr, service.ErrCartNotFound) {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		setInternalServerError(ctx, writer, "Service error")
		return
	}

	if len(cart.Items) == 0 {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	buf, err := json.Marshal(&getCartResponse{
		Items:      cart.Items,
		TotalPrice: cart.TotalPrice,
	})
	if err != nil {
		setInternalServerError(ctx, writer, "Json marshal error")
		return
	}

	writer.WriteHeader(http.StatusOK)

	if _, err = writer.Write(buf); err != nil {
		logger.Errorw(ctx, "Writer error", "error", err)
	}
}
