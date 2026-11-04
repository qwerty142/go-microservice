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

	var parseUserErr error = nil
	var parseSkuErr error = nil
	requestData := &addItemRequest{}

	requestData.UserID, parseUserErr = parseInt64(request.PathValue("user_id"))
	requestData.SkuID, parseSkuErr = parseInt64(request.PathValue("sku_id"))

	if parseUserErr != nil || parseSkuErr != nil {
		setInternalServerError(writer, "Parse query error")
		return
	}

	buf, errRead := io.ReadAll(request.Body)

	defer func(body io.ReadCloser) {
		if closeErr := body.Close(); closeErr != nil {
			logger.Errorw(ctx, "Error closing response body", "error", closeErr)
		}
	}(request.Body)

	if errRead != nil {
		setInternalServerError(writer, "Parse body error")
		return
	}

	if jsonErr := json.Unmarshal(buf, &requestData); jsonErr != nil {
		setInternalServerError(writer, "Unmarshal json error")
		return
	}

	// Validation
	if valid, err := gody.DefaultValidate(*requestData, nil); err != nil || !valid {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	serviceErr := s.cartService.AddItemToCart(ctx, requestData.UserID, requestData.SkuID, requestData.Count)
	if serviceErr != nil {
		var re *products.RequestError
		if ok := errors.As(serviceErr, &re); ok && re.NotFound() {
			writer.WriteHeader(http.StatusPreconditionFailed)
		} else {
			setInternalServerError(writer, "Service error")
		}

		return
	}

	writer.WriteHeader(http.StatusOK)
}
