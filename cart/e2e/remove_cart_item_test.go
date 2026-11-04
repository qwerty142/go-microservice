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
)

func TestPositive_RemoveCartItem(t *testing.T) {
	httpServer := httptest.NewServer(initServer())
	defer httpServer.Close()

	itemData := addItemBody{Count: 10}
	jsonItem, _ := json.Marshal(itemData)
	resp, err := http.Post(httpServer.URL+"/user/31337/cart/1076963", "application/json", bytes.NewBuffer(jsonItem))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	request, err := http.NewRequest(http.MethodDelete, httpUrl+"/user/31337/cart/1076963", bytes.NewBuffer(make([]byte, 0)))
	client := http.Client{}
	respA, errA := client.Do(request)
	require.NoError(t, errA)
	require.Equal(t, http.StatusNoContent, respA.StatusCode)

	resp, err = http.Get(httpUrl + "/user/31337/cart/list")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
