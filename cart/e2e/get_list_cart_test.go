//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"route256/cart/internal/pkg/carts/model"
)

type getCartResponse struct {
	Items      []model.Item `json:"items"`
	TotalPrice uint32       `json:"total_price"`
}

func TestPositive_GetCart(t *testing.T) {
	httpServer := httptest.NewServer(initServer())
	defer httpServer.Close()

	itemData := addItemBody{Count: 10}
	jsonItem, _ := json.Marshal(itemData)
	resp, err := http.Post(httpUrl+"/user/31337/cart/1076963", "application/json", bytes.NewBuffer(jsonItem))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get(httpUrl + "/user/31337/cart/list")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var cart getCartResponse
	_ = json.NewDecoder(resp.Body).Decode(&cart)
	require.Equal(t, 1, len(cart.Items))
	require.Equal(t, model.Item{SkuID: 1076963, Count: 10, Price: 3379, Name: "Теория нравственных чувств | Смит Адам"}, cart.Items[0])
	require.Equal(t, uint32(33790), cart.TotalPrice)
}
