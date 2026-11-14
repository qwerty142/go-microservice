package products

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"route256/cart/internal/pkg/carts/model"
	"route256/cart/internal/pkg/metrics"
	"route256/cart/internal/pkg/util/logger"
)

const getProductPath = "/get_product"

// ProductClientHTTP http client to send requests.
type ProductClientHTTP struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewProductClientHTTP constructor for ProductClientHTTP.
func NewProductClientHTTP(baseURL, token string, client *http.Client) *ProductClientHTTP {
	return &ProductClientHTTP{baseURL: baseURL, token: token, client: client}
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

// GetProduct fetches model.ProductInfo from remote server.
func (c *ProductClientHTTP) GetProduct(ctx context.Context, skuID int64) (model.ProductInfo, error) {
	requestBody, err := json.Marshal(getProductRequest{Token: c.token, SKU: skuID})
	if err != nil {
		return model.ProductInfo{}, fmt.Errorf("marshal get_product request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+getProductPath, bytes.NewReader(requestBody))
	if err != nil {
		return model.ProductInfo{}, fmt.Errorf("create get_product request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	startTime := time.Now()
	response, err := c.client.Do(req)
	if err != nil {
		return model.ProductInfo{}, fmt.Errorf("send get_product request: %w", err)
	}
	metrics.SaveExternalCallDuration(startTime, getProductPath, response.StatusCode)

	defer func(body io.ReadCloser) {
		if closeErr := body.Close(); closeErr != nil {
			logger.Errorw(ctx, "Error closing response body", "error", closeErr)
		}
	}(response.Body)

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return model.ProductInfo{}, fmt.Errorf("read get_product response body: %w", err)
	}

	if response.StatusCode == http.StatusOK {
		productResp := productResponse{}
		if err = json.Unmarshal(responseBody, &productResp); err != nil {
			return model.ProductInfo{}, fmt.Errorf("unmarshal get_product response: %w", err)
		}

		return model.ProductInfo{Name: productResp.Name, Price: productResp.Price}, nil
	}

	errorResp := errorResponse{}
	if err = json.Unmarshal(responseBody, &errorResp); err != nil {
		errorResp.Message = response.Status
	}

	return model.ProductInfo{}, &RequestError{StatusCode: response.StatusCode, Err: errors.New(errorResp.Message)}
}
