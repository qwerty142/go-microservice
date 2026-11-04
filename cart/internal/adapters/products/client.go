package products

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"route256/cart/internal/pkg/carts/model"
	"route256/cart/internal/pkg/metrics"
	"route256/cart/internal/pkg/util/logger"
)

// ProductClientHTTP http client to send requests.
type ProductClientHTTP struct {
	token  string
	client *http.Client
}

// NewProductClientHTTP constructor for ProductClientHTTP.
func NewProductClientHTTP(token string, client *http.Client) *ProductClientHTTP {
	return &ProductClientHTTP{token, client}
}

type getProductRequest struct {
	Token string `json:"token"`
	SKU   int64  `json:"sku"`
}

type productResponse struct {
	Name  string `json:"name"`
	Price uint32 `json:"price"`
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// GetProduct fetches service.ProductInfo from remote server.
func (c *ProductClientHTTP) GetProduct(ctx context.Context, skuID int64) (model.ProductInfo, error) {
	request := getProductRequest{Token: c.token, SKU: skuID}

	requestBody, err := json.Marshal(request)
	if err != nil {
		logger.Errorw(ctx, "Error marshaling JSON", "error", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"http://route256.pavl.uk:8080/get_product",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		logger.Errorw(ctx, "Error creating request", "error", err)
	}

	req.Header.Set("Content-Type", "application/json")

	startTime := time.Now()
	response, err := c.client.Do(req)
	metrics.SaveExternalCallDuration(startTime, "/get_product", response.StatusCode)

	if err != nil {
		logger.Errorw(ctx, "Error sending request", "error", err)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Errorw(ctx, "Error reading response body", "error", err)
	}

	defer func(body io.ReadCloser) {
		if closeErr := body.Close(); closeErr != nil {
			logger.Errorw(ctx, "Error closing response body", "error", closeErr)
		}
	}(response.Body)

	if response.StatusCode == http.StatusOK {
		productResp := productResponse{}
		if err = json.Unmarshal(responseBody, &productResp); err != nil {
			logger.Errorw(ctx, "Error unmarshaling response body", "error", err)
		}

		return model.ProductInfo{Name: productResp.Name, Price: productResp.Price}, nil
	}

	errorResp := errorResponse{}
	if err = json.Unmarshal(responseBody, &errorResp); err != nil {
		logger.Errorw(ctx, "Error unmarshaling response body", "error", err)
	}

	return model.ProductInfo{}, &RequestError{StatusCode: response.StatusCode, Err: errors.New(errorResp.Message)}
}
