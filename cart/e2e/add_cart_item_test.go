//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type addItemBody struct {
	Count uint16 `json:"count"`
}

func TestPositive_AddItem(t *testing.T) {
	httpServer := httptest.NewServer(initServer())
	defer httpServer.Close()

	itemData := addItemBody{Count: 123}
	jsonItem, _ := json.Marshal(itemData)
	resp, err := http.Post(httpUrl+"/user/31337/cart/1076963", "application/json", bytes.NewBuffer(jsonItem))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
